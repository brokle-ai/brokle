package evaluation

import (
	evaluationDomain "brokle/internal/core/domain/evaluation"
)

// Huma operation types for the wizard feature of the evaluation package.

type WizardCreateInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	Body      evaluationDomain.CreateExperimentFromWizardRequest
}

type WizardValidateInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	Body      evaluationDomain.ValidateStepRequest
}

type WizardValidateOutput struct {
	Body *evaluationDomain.ValidateStepResponse
}

type WizardEstimateInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	Body      evaluationDomain.EstimateCostRequest
}

type WizardEstimateOutput struct {
	Body *evaluationDomain.EstimateCostResponse
}

type WizardDatasetFieldsInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	DatasetID string `path:"datasetId" format:"uuid"`
}

type WizardDatasetFieldsOutput struct {
	Body *evaluationDomain.DatasetFieldsResponse
}

type WizardGetConfigInput struct {
	ProjectID    string `path:"projectId" format:"uuid"`
	ExperimentID string `path:"experimentId" format:"uuid"`
}

type WizardGetConfigOutput struct {
	Body *evaluationDomain.ExperimentConfigResponse
}
