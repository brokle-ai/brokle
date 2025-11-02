/**
 * Onboarding State Reducer
 * Manages onboarding flow state with type-safe actions
 */ 

import type { OnboardingState, OnboardingAction } from './onboardingTypes'

export const initialOnboardingState: OnboardingState = {
  currentStep: 0,
  direction: 'forward',
  isSubmitting: false,
}

export function onboardingReducer(
  state: OnboardingState,
  action: OnboardingAction
): OnboardingState {
  switch (action.type) {
    case 'next':
      return {
        ...state,
        currentStep: state.currentStep + 1,
        direction: 'forward',
      }

    case 'back':
      return {
        ...state,
        currentStep: Math.max(0, state.currentStep - 1),
        direction: 'backward',
      }

    case 'goToStep':
      return {
        ...state,
        currentStep: action.step,
        direction: action.step > state.currentStep ? 'forward' : 'backward',
      }

    case 'setSubmitting':
      return {
        ...state,
        isSubmitting: action.isSubmitting,
      }

    default:
      return state
  }
}
