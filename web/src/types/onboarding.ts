export interface OnboardingQuestion {
  id: string
  step: number
  question_type: 'single_choice' | 'multiple_choice' | 'text'
  title: string
  description?: string
  is_required: boolean
  options?: string[]
  user_answer?: any
  is_skipped: boolean
}

export interface OnboardingStatus {
  total_questions: number
  completed_questions: number
  skipped_questions: number
  remaining_questions: number
  onboarding_completed: boolean // Computed from backend
  current_step: number
}

export interface SubmitResponseRequest {
  question_id: string
  response_value?: any
  skipped: boolean
}

export interface SubmitResponsesRequest {
  responses: SubmitResponseRequest[]
}

export interface OnboardingApiResponse<T = any> {
  success: boolean
  data: T
  meta: {
    request_id: string
  }
}

export type QuestionType = OnboardingQuestion['question_type']

// UI State interfaces
export interface OnboardingState {
  questions: OnboardingQuestion[]
  currentStep: number
  status: OnboardingStatus | null
  loading: boolean
  error: string | null
  submitting: boolean
}

export interface QuestionResponse {
  questionId: string
  value: any
  skipped: boolean
}