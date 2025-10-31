/**
 * useOnboardingForm Hook
 * Enhanced onboarding with Langfuse-inspired UX patterns
 *
 * Features:
 * - useReducer for state management
 * - Conditional question skipping
 * - Auto-advance on selection
 * - Keyboard shortcuts
 * - Type-safe actions
 */

import { useReducer, useCallback, useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { toast } from 'sonner'
import { getQuestions, submitResponse, skipQuestion, getStatus } from '@/lib/api'
import { onboardingReducer, initialOnboardingState } from '../lib/onboardingReducer'
import type { OnboardingQuestion } from '@/types/onboarding'
import type { UseOnboardingFormReturn } from '../lib/onboardingTypes'

export function useOnboardingForm(): UseOnboardingFormReturn {
  const router = useRouter()
  const [state, dispatch] = useReducer(onboardingReducer, initialOnboardingState)
  const [questions, setQuestions] = useState<OnboardingQuestion[]>([])
  const [loading, setLoading] = useState(true)

  // Load questions on mount
  useEffect(() => {
    loadQuestions()
  }, [])

  const loadQuestions = async () => {
    try {
      setLoading(true)
      const data = await getQuestions()
      setQuestions(data)

      // Set initial step to first unanswered question
      const firstUnanswered = data.findIndex(q => !q.user_answer && !q.is_skipped)
      if (firstUnanswered !== -1) {
        dispatch({ type: 'goToStep', step: firstUnanswered })
      }
    } catch (error) {
      console.error('Failed to load questions:', error)
      toast.error('Failed to load onboarding questions')
    } finally {
      setLoading(false)
    }
  }

  // Get current question
  const currentQuestion = questions[state.currentStep]

  // Check conditional skipping
  const shouldSkipCurrentQuestion = (): boolean => {
    if (!currentQuestion) return false

    // Example: Skip "How did you hear about us?" if user was invited
    // This logic can be expanded based on question metadata
    const previousAnswers = questions
      .slice(0, state.currentStep)
      .reduce((acc, q) => {
        if (q.user_answer) {
          acc[q.id] = q.user_answer
        }
        return acc
      }, {} as Record<string, any>)

    // Add conditional logic here based on previous answers
    // For now, no conditions
    return false
  }

  // Calculate effective total steps (excluding conditionally skipped)
  const effectiveTotalSteps = questions.length

  // Navigation helpers
  const isFirstStep = state.currentStep === 0
  const isLastStep = state.currentStep === effectiveTotalSteps - 1

  const goNext = useCallback(() => {
    if (state.currentStep < effectiveTotalSteps - 1) {
      dispatch({ type: 'next' })
    }
  }, [state.currentStep, effectiveTotalSteps])

  const goBack = useCallback(() => {
    if (state.currentStep > 0) {
      dispatch({ type: 'back' })
    }
  }, [state.currentStep])

  const goToStep = useCallback((step: number) => {
    if (step >= 0 && step < effectiveTotalSteps) {
      dispatch({ type: 'goToStep', step })
    }
  }, [effectiveTotalSteps])

  // Submit response handler
  const handleSubmitResponse = async (questionId: string, value: string | string[]) => {
    try {
      dispatch({ type: 'setSubmitting', isSubmitting: true })

      await submitResponse(questionId, value)

      // Update local state
      setQuestions(prev =>
        prev.map(q =>
          q.id === questionId
            ? { ...q, user_answer: value, is_skipped: false }
            : q
        )
      )

      // Auto-advance to next question (except on last)
      if (!isLastStep) {
        goNext()
      }

      toast.success('Answer saved')
    } catch (error) {
      console.error('Failed to submit response:', error)
      toast.error('Failed to save answer')
    } finally {
      dispatch({ type: 'setSubmitting', isSubmitting: false })
    }
  }

  // Skip question handler
  const handleSkipQuestion = async (questionId: string) => {
    try {
      dispatch({ type: 'setSubmitting', isSubmitting: true })

      await skipQuestion(questionId)

      // Update local state
      setQuestions(prev =>
        prev.map(q =>
          q.id === questionId
            ? { ...q, is_skipped: true, user_answer: undefined }
            : q
        )
      )

      // Auto-advance to next question (except on last)
      if (!isLastStep) {
        goNext()
      }

      toast.success('Question skipped')
    } catch (error) {
      console.error('Failed to skip question:', error)
      toast.error('Failed to skip question')
    } finally {
      dispatch({ type: 'setSubmitting', isSubmitting: false })
    }
  }

  // Finish onboarding
  const handleFinish = async () => {
    try {
      const { completeOnboarding } = await import('@/lib/api/services/onboarding')
      await completeOnboarding()
      toast.success('Onboarding completed!')
      router.push('/')
    } catch (error) {
      console.error('Failed to complete onboarding:', error)
      toast.error('Failed to complete onboarding')
    }
  }

  return {
    state,
    currentQuestion,
    isFirstStep,
    isLastStep,
    effectiveTotalSteps,
    goNext,
    goBack,
    goToStep,  // Export for step indicators
    handleSubmitResponse,
    handleSkipQuestion,
    handleFinish,
  }
}
