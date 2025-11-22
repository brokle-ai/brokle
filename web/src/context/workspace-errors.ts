/**
 * Workspace error classification system
 * Provides specific, actionable error messages for workspace operations
 */

export enum WorkspaceErrorCode {
  // Slug validation errors
  INVALID_ORG_SLUG = 'INVALID_ORG_SLUG',
  INVALID_PROJECT_SLUG = 'INVALID_PROJECT_SLUG',

  // Not found errors
  ORG_NOT_FOUND = 'ORG_NOT_FOUND',
  PROJECT_NOT_FOUND = 'PROJECT_NOT_FOUND',

  // Access errors
  ORG_NO_ACCESS = 'ORG_NO_ACCESS',
  PROJECT_NO_ACCESS = 'PROJECT_NO_ACCESS',

  // API errors
  API_FAILED = 'API_FAILED',
  NETWORK_ERROR = 'NETWORK_ERROR',

  // Generic fallback
  UNKNOWN = 'UNKNOWN',
}

/**
 * WorkspaceError class for type-safe error handling
 * Enables instanceof checks to preserve already-classified errors
 *
 * Separates developer messages (technical details for logs) from user messages (friendly UI text)
 */
export class WorkspaceError extends Error {
  constructor(
    public readonly code: WorkspaceErrorCode,
    message: string,  // Developer/technical message (stored in Error.message)
    public readonly userMessage: string,  // User-facing message for UI
    public readonly context?: Record<string, unknown>,
    public readonly originalError?: Error
  ) {
    super(message)  // Pass technical message to Error.message for developer logs
    this.name = 'WorkspaceError'

    // Maintain proper prototype chain for instanceof checks
    Object.setPrototypeOf(this, WorkspaceError.prototype)
  }
}

/**
 * Create a workspace error with proper classification
 */
export function createWorkspaceError(
  code: WorkspaceErrorCode,
  context?: Record<string, unknown>,
  originalError?: Error
): WorkspaceError {
  const errorMap: Record<WorkspaceErrorCode, { message: string; userMessage: string }> = {
    [WorkspaceErrorCode.INVALID_ORG_SLUG]: {
      message: 'Invalid organization slug format',
      userMessage: 'The organization link is invalid. Please check the URL or contact support.',
    },
    [WorkspaceErrorCode.INVALID_PROJECT_SLUG]: {
      message: 'Invalid project slug format',
      userMessage: 'The project link is invalid. Please check the URL or contact support.',
    },
    [WorkspaceErrorCode.ORG_NOT_FOUND]: {
      message: 'Organization not found in user workspace',
      userMessage: "This organization doesn't exist or you don't have access to it.",
    },
    [WorkspaceErrorCode.PROJECT_NOT_FOUND]: {
      message: 'Project not found in organization',
      userMessage: "This project doesn't exist or has been deleted.",
    },
    [WorkspaceErrorCode.ORG_NO_ACCESS]: {
      message: 'User does not have access to organization',
      userMessage: "You don't have permission to access this organization.",
    },
    [WorkspaceErrorCode.PROJECT_NO_ACCESS]: {
      message: 'User does not have access to project',
      userMessage: "You don't have permission to access this project.",
    },
    [WorkspaceErrorCode.API_FAILED]: {
      message: 'API request failed',
      userMessage: 'Something went wrong. Please try again.',
    },
    [WorkspaceErrorCode.NETWORK_ERROR]: {
      message: 'Network connection failed',
      userMessage: 'Unable to connect. Please check your internet connection.',
    },
    [WorkspaceErrorCode.UNKNOWN]: {
      message: 'Unknown error occurred',
      userMessage: 'An unexpected error occurred. Please try again.',
    },
  }

  const errorInfo = errorMap[code]

  return new WorkspaceError(
    code,
    errorInfo.message,      // Developer/technical message
    errorInfo.userMessage,  // User-facing message
    context,
    originalError
  )
}

/**
 * Detect error type from slug validation
 */
export function classifySlugError(
  slug: string,
  type: 'organization' | 'project'
): WorkspaceError {
  const code = type === 'organization'
    ? WorkspaceErrorCode.INVALID_ORG_SLUG
    : WorkspaceErrorCode.INVALID_PROJECT_SLUG

  return createWorkspaceError(code, { slug, type })
}

/**
 * Detect error type from API error
 */
export function classifyAPIError(error: unknown): WorkspaceError {
  // Type guard for errors with statusCode (BrokleAPIError)
  const hasStatusCode = (err: unknown): err is { statusCode: number } & Error =>
    typeof err === 'object' && err !== null && 'statusCode' in err

  // Type guard for network errors
  const hasNetworkError = (err: unknown): err is { isNetworkError: () => boolean } & Error =>
    typeof err === 'object' && err !== null && 'isNetworkError' in err

  // Check if it's a BrokleAPIError
  if (hasStatusCode(error)) {
    if (error.statusCode === 404) {
      return createWorkspaceError(WorkspaceErrorCode.ORG_NOT_FOUND, {}, error)
    }
    if (error.statusCode === 403) {
      return createWorkspaceError(WorkspaceErrorCode.ORG_NO_ACCESS, {}, error)
    }
    if (error.statusCode >= 500) {
      return createWorkspaceError(WorkspaceErrorCode.API_FAILED, {}, error)
    }
  }

  // Network error
  if (hasNetworkError(error) && error.isNetworkError()) {
    return createWorkspaceError(WorkspaceErrorCode.NETWORK_ERROR, {}, error)
  }

  // Fallback
  const errorObj = error instanceof Error ? error : undefined
  return createWorkspaceError(WorkspaceErrorCode.UNKNOWN, {}, errorObj)
}
