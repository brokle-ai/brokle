// Package registration orchestrates user registration across auth, organization, and user domains.
package registration

import (
	"context"
	"time"

	"golang.org/x/crypto/bcrypt"

	authDomain "brokle/internal/core/domain/auth"
	"brokle/internal/core/domain/common"
	orgDomain "brokle/internal/core/domain/organization"
	userDomain "brokle/internal/core/domain/user"
	auth "brokle/internal/core/services/auth"
	billingService "brokle/internal/core/services/billing"
	appErrors "brokle/pkg/errors"
)

// RegisterRequest contains all data needed for registration
type RegisterRequest struct {
	ReferralSource   *string
	OrganizationName *string
	InvitationToken  *string
	Email            string
	Password         string
	FirstName        string
	LastName         string
	Role             string
	Provider         string
	ProviderID       string
	IsOAuthUser      bool
}

// OAuthRegistrationRequest contains OAuth-specific registration data
type OAuthRegistrationRequest struct {
	ReferralSource   *string
	OrganizationName *string
	InvitationToken  *string
	Email            string
	FirstName        string
	LastName         string
	Role             string
	Provider         string
	ProviderID       string
}

// RegistrationResponse contains the result of a successful registration
type RegistrationResponse struct {
	User         *userDomain.User
	Organization *orgDomain.Organization
	Project      *orgDomain.Project // Will be nil for invitation-based signup
	LoginTokens  *authDomain.LoginResponse
}

type RegistrationService struct {
	transactor           common.Transactor
	userRepo             userDomain.Repository
	orgRepo              orgDomain.OrganizationRepository
	memberRepo           orgDomain.MemberRepository
	projectRepo          orgDomain.ProjectRepository
	invitationRepo       orgDomain.InvitationRepository
	roles          *auth.RoleService
	auth          *auth.AuthService
	usage *billingService.BillableUsageService
}

// NewRegistrationService creates a new registration service
func NewRegistrationService(
	transactor common.Transactor,
	userRepo userDomain.Repository,
	orgRepo orgDomain.OrganizationRepository,
	memberRepo orgDomain.MemberRepository,
	projectRepo orgDomain.ProjectRepository,
	invitationRepo orgDomain.InvitationRepository,
	roles *auth.RoleService,
	auth *auth.AuthService,
	usage *billingService.BillableUsageService,
) *RegistrationService {
	return &RegistrationService{
		transactor:           transactor,
		userRepo:             userRepo,
		orgRepo:              orgRepo,
		memberRepo:           memberRepo,
		projectRepo:          projectRepo,
		invitationRepo:       invitationRepo,
		roles:          roles,
		auth:          auth,
		usage: usage,
	}
}

// RegisterWithOrganization handles fresh signup: user + organization + project
func (s *RegistrationService) RegisterWithOrganization(ctx context.Context, req *RegisterRequest) (*RegistrationResponse, error) {
	// Validation
	if req.OrganizationName == nil || *req.OrganizationName == "" {
		return nil, appErrors.InvalidParam("organization_name", "is required")
	}

	if !req.IsOAuthUser && req.Password == "" {
		return nil, appErrors.InvalidParam("password", "is required for email signups")
	}

	// Hash password BEFORE transaction (don't re-hash inside)
	hashedPassword, err := s.hashPassword(req.Password, req.IsOAuthUser)
	if err != nil {
		return nil, err
	}

	var newUser *userDomain.User
	var org *orgDomain.Organization
	var project *orgDomain.Project

	// TRANSACTION: Create user, org, project, and membership atomically
	err = s.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		// 1. Create user
		newUser = userDomain.NewUser(req.Email, req.FirstName, req.LastName, req.Role)
		newUser.ReferralSource = req.ReferralSource

		// Set authentication method and OAuth provider info
		if req.IsOAuthUser {
			newUser.AuthMethod = "oauth"
			newUser.OAuthProvider = &req.Provider
			newUser.OAuthProviderID = &req.ProviderID
			newUser.Password = nil // NULL password for OAuth users

			// OAuth emails are pre-verified by provider
			now := time.Now()
			newUser.IsEmailVerified = true
			newUser.EmailVerifiedAt = &now
		} else {
			newUser.AuthMethod = "password"
			newUser.SetPassword(hashedPassword) // Reuse pre-hashed password
			newUser.IsEmailVerified = false     // Email/password users need verification
		}

		if err := s.userRepo.Create(ctx, newUser); err != nil {
			// Repo wraps UNIQUE (email) collisions into appErrors.IsAlreadyExists.
			if appErrors.IsAlreadyExists(err) {
				return appErrors.Conflict("", "email already registered")
			}
			return appErrors.Internal("failed to create user", err)
		}

		// Create user profile
		profile := userDomain.NewUserProfile(newUser.ID)
		if err := s.userRepo.CreateProfile(ctx, profile); err != nil {
			return appErrors.Internal("failed to create user profile", err)
		}

		// 2. Create organization
		org = orgDomain.NewOrganization(*req.OrganizationName)
		org.BillingEmail = &req.Email
		org.Plan = "free"
		org.SubscriptionStatus = "active"

		if err := s.orgRepo.Create(ctx, org); err != nil {
			return appErrors.Internal("failed to create organization", err)
		}

		// 2.5. Provision billing for organization
		if err := s.usage.ProvisionOrganizationBilling(ctx, org.ID); err != nil {
			return appErrors.Internal("failed to provision billing", err)
		}

		// 3. Add user as organization owner
		ownerRole, err := s.roles.GetRoleByNameAndScope(ctx, "owner", "organization")
		if err != nil || ownerRole == nil {
			return appErrors.Internal("owner role not found - database seed may be missing", err)
		}

		member := orgDomain.NewMember(org.ID, newUser.ID, ownerRole.ID)
		if err := s.memberRepo.Create(ctx, member); err != nil {
			return appErrors.Internal("failed to add user as organization owner", err)
		}

		// 4. Create default project
		project = orgDomain.NewProject(org.ID, "Default Project", "Your default project")
		if err := s.projectRepo.Create(ctx, project); err != nil {
			return appErrors.Internal("failed to create default project", err)
		}

		// 5. Set user's default organization
		newUser.DefaultOrganizationID = &org.ID
		if err := s.userRepo.Update(ctx, newUser); err != nil {
			return appErrors.Internal("failed to set default organization", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Generate login tokens (OUTSIDE transaction)
	var loginTokens *authDomain.LoginResponse
	if req.IsOAuthUser {
		// OAuth users: generate tokens without password validation
		loginTokens, err = s.auth.GenerateTokensForUser(ctx, newUser.ID)
		if err != nil {
			// Non-critical - user created successfully, just can't auto-login
			loginTokens = nil
		}
	} else {
		// Email/password users: use Login with password validation
		loginReq := &authDomain.LoginRequest{
			Email:    req.Email,
			Password: req.Password,
		}
		loginTokens, err = s.auth.Login(ctx, loginReq)
		if err != nil {
			// Non-critical - user created successfully, just can't auto-login
			loginTokens = nil
		}
	}

	return &RegistrationResponse{
		User:         newUser,
		Organization: org,
		Project:      project,
		LoginTokens:  loginTokens,
	}, nil
}

// RegisterWithInvitation handles invitation-based signup
func (s *RegistrationService) RegisterWithInvitation(ctx context.Context, req *RegisterRequest) (*RegistrationResponse, error) {
	// Validation
	if req.InvitationToken == nil || *req.InvitationToken == "" {
		return nil, appErrors.InvalidParam("invitation_token", "is required")
	}

	if !req.IsOAuthUser && req.Password == "" {
		return nil, appErrors.InvalidParam("password", "is required for email signups")
	}

	// Get invitation
	invitation, err := s.invitationRepo.GetByToken(ctx, *req.InvitationToken)
	if err != nil {
		return nil, appErrors.NotFound("invitation", appErrors.WithMessage("invalid invitation token"))
	}

	// Check if expired
	if invitation.Status != orgDomain.InvitationStatusPending || time.Now().After(invitation.ExpiresAt) {
		return nil, appErrors.InvalidParam("invitation_token", "invitation has expired or is no longer valid")
	}

	// Verify email matches invitation
	if invitation.Email != req.Email {
		return nil, appErrors.InvalidParam("email", "does not match invitation")
	}

	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil && !appErrors.IsNotFound(err) {
		return nil, appErrors.Internal("user lookup failed", err)
	}
	if existingUser != nil {
		return nil, appErrors.Conflict("", "email already exists")
	}

	// Hash password BEFORE transaction (don't re-hash inside)
	hashedPassword, err := s.hashPassword(req.Password, req.IsOAuthUser)
	if err != nil {
		return nil, err
	}

	var newUser *userDomain.User
	var org *orgDomain.Organization

	// TRANSACTION: Create user and accept invitation
	err = s.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		// 1. Create user
		newUser = userDomain.NewUser(req.Email, req.FirstName, req.LastName, req.Role)
		newUser.ReferralSource = req.ReferralSource

		// Set authentication method and OAuth provider info (same logic as RegisterWithOrganization)
		if req.IsOAuthUser {
			newUser.AuthMethod = "oauth"
			newUser.OAuthProvider = &req.Provider
			newUser.OAuthProviderID = &req.ProviderID
			newUser.Password = nil // NULL password for OAuth users

			// OAuth emails are pre-verified by provider
			now := time.Now()
			newUser.IsEmailVerified = true
			newUser.EmailVerifiedAt = &now
		} else {
			newUser.AuthMethod = "password"
			newUser.SetPassword(hashedPassword) // Reuse pre-hashed password
			newUser.IsEmailVerified = false
		}

		// Set default organization to invited org
		newUser.DefaultOrganizationID = &invitation.OrganizationID

		if err := s.userRepo.Create(ctx, newUser); err != nil {
			// Repo wraps UNIQUE (email) collisions into appErrors.IsAlreadyExists.
			if appErrors.IsAlreadyExists(err) {
				return appErrors.Conflict("", "email already registered")
			}
			return appErrors.Internal("failed to create user", err)
		}

		// Create user profile (same as fresh signup)
		profile := userDomain.NewUserProfile(newUser.ID)
		if err := s.userRepo.CreateProfile(ctx, profile); err != nil {
			return appErrors.Internal("failed to create user profile", err)
		}

		// 2. Update invitation status to accepted
		invitation.Status = orgDomain.InvitationStatusAccepted
		acceptedAt := time.Now()
		invitation.AcceptedAt = &acceptedAt
		if err := s.invitationRepo.Update(ctx, invitation); err != nil {
			return appErrors.Internal("failed to update invitation", err)
		}

		// 3. Add user as organization member
		member := orgDomain.NewMember(invitation.OrganizationID, newUser.ID, invitation.RoleID)
		if err := s.memberRepo.Create(ctx, member); err != nil {
			return appErrors.Internal("failed to add user to organization", err)
		}

		// 4. Get organization details
		org, err = s.orgRepo.GetByID(ctx, invitation.OrganizationID)
		if err != nil {
			return appErrors.Internal("failed to get organization", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Generate login tokens (OUTSIDE transaction)
	var loginTokens *authDomain.LoginResponse
	if req.IsOAuthUser {
		// OAuth users: generate tokens without password validation
		loginTokens, err = s.auth.GenerateTokensForUser(ctx, newUser.ID)
		if err != nil {
			loginTokens = nil
		}
	} else {
		// Email/password users: use Login with password validation
		loginReq := &authDomain.LoginRequest{
			Email:    req.Email,
			Password: req.Password,
		}
		loginTokens, err = s.auth.Login(ctx, loginReq)
		if err != nil {
			loginTokens = nil
		}
	}

	return &RegistrationResponse{
		User:         newUser,
		Organization: org,
		Project:      nil, // No project created for invitation signup
		LoginTokens:  loginTokens,
	}, nil
}

// CompleteOAuthRegistration handles OAuth-based registration
func (s *RegistrationService) CompleteOAuthRegistration(ctx context.Context, req *OAuthRegistrationRequest) (*RegistrationResponse, error) {
	// Convert to RegisterRequest
	regReq := &RegisterRequest{
		Email:            req.Email,
		Password:         "", // No password for OAuth
		FirstName:        req.FirstName,
		LastName:         req.LastName,
		Role:             req.Role,
		ReferralSource:   req.ReferralSource,
		OrganizationName: req.OrganizationName,
		InvitationToken:  req.InvitationToken,
		IsOAuthUser:      true,
		Provider:         req.Provider,
		ProviderID:       req.ProviderID,
	}

	// Route to appropriate registration method
	if req.InvitationToken != nil {
		return s.RegisterWithInvitation(ctx, regReq)
	}
	return s.RegisterWithOrganization(ctx, regReq)
}

// hashPassword handles password hashing for regular users
// OAuth users don't use this - they get NULL password and auth_method='oauth'
func (s *RegistrationService) hashPassword(password string, isOAuthUser bool) (string, error) {
	if isOAuthUser {
		// OAuth users: return empty string (will be set to NULL in database)
		return "", nil
	}

	// Password users: hash the password
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", appErrors.Internal("failed to hash password", err)
	}
	return string(hashed), nil
}
