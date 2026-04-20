package evaluation

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func registerWizardRoutes(api huma.API, h *handler) {
	huma.Register(api, huma.Operation{
		OperationID:   "create-experiment-from-wizard",
		Method:        http.MethodPost,
		Path:          "/api/v1/projects/{projectId}/experiments/wizard",
		Tags:          []string{"experiment-wizard"},
		Summary:       "Create an experiment from the wizard",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.wizardCreate)

	huma.Register(api, huma.Operation{
		OperationID: "validate-experiment-wizard-step",
		Method:      http.MethodPost,
		Path:        "/api/v1/projects/{projectId}/experiments/wizard/validate",
		Tags:        []string{"experiment-wizard"},
		Summary:     "Validate a wizard step",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.wizardValidate)

	huma.Register(api, huma.Operation{
		OperationID: "estimate-experiment-cost",
		Method:      http.MethodPost,
		Path:        "/api/v1/projects/{projectId}/experiments/wizard/estimate",
		Tags:        []string{"experiment-wizard"},
		Summary:     "Estimate experiment cost",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.wizardEstimate)

	huma.Register(api, huma.Operation{
		OperationID: "get-dataset-fields",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/datasets/{datasetId}/fields",
		Tags:        []string{"experiment-wizard"},
		Summary:     "Get dataset field schema for variable mapping",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.wizardDatasetFields)

	huma.Register(api, huma.Operation{
		OperationID: "get-experiment-config",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/experiments/{experimentId}/config",
		Tags:        []string{"experiment-wizard"},
		Summary:     "Get the wizard config for an experiment",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.wizardGetConfig)
}

func (h *handler) wizardCreate(ctx context.Context, in *WizardCreateInput) (*ExperimentOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	body := in.Body
	exp, err := h.experimentWizardSvc.CreateFromWizard(ctx, projectID, userIDPtr(ctx), &body)
	if err != nil {
		return nil, err
	}
	return &ExperimentOutput{Body: exp.ToResponse()}, nil
}

func (h *handler) wizardValidate(ctx context.Context, in *WizardValidateInput) (*WizardValidateOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	body := in.Body
	res, err := h.experimentWizardSvc.ValidateStep(ctx, projectID, &body)
	if err != nil {
		return nil, err
	}
	return &WizardValidateOutput{Body: res}, nil
}

func (h *handler) wizardEstimate(ctx context.Context, in *WizardEstimateInput) (*WizardEstimateOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	body := in.Body
	res, err := h.experimentWizardSvc.EstimateCost(ctx, projectID, &body)
	if err != nil {
		return nil, err
	}
	return &WizardEstimateOutput{Body: res}, nil
}

func (h *handler) wizardDatasetFields(ctx context.Context, in *WizardDatasetFieldsInput) (*WizardDatasetFieldsOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	datasetID, err := parseDatasetID(in.DatasetID)
	if err != nil {
		return nil, err
	}
	res, err := h.experimentWizardSvc.GetDatasetFields(ctx, projectID, datasetID)
	if err != nil {
		return nil, err
	}
	return &WizardDatasetFieldsOutput{Body: res}, nil
}

func (h *handler) wizardGetConfig(ctx context.Context, in *WizardGetConfigInput) (*WizardGetConfigOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	experimentID, err := parseExperimentID(in.ExperimentID)
	if err != nil {
		return nil, err
	}
	cfg, err := h.experimentWizardSvc.GetExperimentConfig(ctx, experimentID, projectID)
	if err != nil {
		return nil, err
	}
	return &WizardGetConfigOutput{Body: cfg.ToResponse()}, nil
}
