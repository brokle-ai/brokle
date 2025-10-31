/**
 * Onboarding Types
 * Type-safe state management for onboarding flow
 */

import type { OnboardingQuestion } from '@/types/onboarding'

export interface OnboardingFormData {
  [questionId: string]: string | string[] | undefined
}

export interface OnboardingState {
  currentStep: number
  direction: 'forward' | 'backward'
  isSubmitting: boolean
}

export type OnboardingAction =
  | { type: 'next' }
  | { type: 'back' }
  | { type: 'goToStep'; step: number }
  | { type: 'setSubmitting'; isSubmitting: boolean }

export interface UseOnboardingFormReturn {
  // State
  state: OnboardingState
  currentQuestion: OnboardingQuestion | undefined
  isFirstStep: boolean
  isLastStep: boolean
  effectiveTotalSteps: number

  // Actions
  goNext: () => void
  goBack: () => void
  goToStep: (step: number) => void

  // Form
  handleSubmitResponse: (questionId: string, value: string | string[]) => Promise<void>
  handleSkipQuestion: (questionId: string) => Promise<void>
  handleFinish: () => void
}
