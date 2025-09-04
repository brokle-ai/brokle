'use client'

import { Button } from '@/components/ui/button'
import { ChevronLeft, ChevronRight, CheckCircle } from 'lucide-react'
import type { OnboardingQuestion, OnboardingStatus } from '@/types/onboarding'

interface OnboardingNavigationProps {
  questions: OnboardingQuestion[]
  currentStep: number
  status: OnboardingStatus | null
  onPrevious: () => void
  onNext: () => void
  onComplete: () => void
  isSubmitting?: boolean
}

export function OnboardingNavigation({
  questions,
  currentStep,
  status,
  onPrevious,
  onNext,
  onComplete,
  isSubmitting = false
}: OnboardingNavigationProps) {
  if (!status || questions.length === 0) {
    return null
  }

  const currentQuestion = questions.find(q => q.step === currentStep)
  const isFirstQuestion = currentStep === 1
  const isLastQuestion = currentStep === status.total_questions
  const canGoNext = currentQuestion && (
    currentQuestion.user_answer !== undefined || 
    currentQuestion.is_skipped
  )
  
  // Check if onboarding is complete
  const isComplete = status.onboarding_completed
  
  return (
    <div className="flex items-center justify-between w-full p-4 bg-background border-t border-border">
      {/* Previous Button */}
      <div className="flex-1">
        {!isFirstQuestion && (
          <Button
            variant="outline"
            onClick={onPrevious}
            disabled={isSubmitting}
            className="flex items-center gap-2"
          >
            <ChevronLeft className="h-4 w-4" />
            Previous
          </Button>
        )}
      </div>

      {/* Step Indicator */}
      <div className="flex-1 text-center">
        <div className="text-sm text-muted-foreground">
          Question {currentStep} of {status.total_questions}
        </div>
        <div className="text-xs text-muted-foreground mt-1">
          {status.completed_questions} answered • {status.skipped_questions} skipped
        </div>
      </div>

      {/* Next/Complete Button */}
      <div className="flex-1 flex justify-end">
        {isComplete ? (
          <Button
            onClick={onComplete}
            disabled={isSubmitting}
            className="flex items-center gap-2 bg-green-600 hover:bg-green-700"
          >
            <CheckCircle className="h-4 w-4" />
            Complete Setup
          </Button>
        ) : isLastQuestion && canGoNext ? (
          <Button
            onClick={onNext}
            disabled={isSubmitting || !canGoNext}
            className="flex items-center gap-2"
          >
            Finish
            <CheckCircle className="h-4 w-4" />
          </Button>
        ) : (
          <Button
            variant={canGoNext ? "default" : "outline"}
            onClick={onNext}
            disabled={isSubmitting || !canGoNext}
            className="flex items-center gap-2"
          >
            Next
            <ChevronRight className="h-4 w-4" />
          </Button>
        )}
      </div>
    </div>
  )
}

// Compact navigation for mobile/smaller screens
export function CompactOnboardingNavigation({
  questions,
  currentStep,
  status,
  onPrevious,
  onNext,
  onComplete,
  isSubmitting = false
}: OnboardingNavigationProps) {
  if (!status || questions.length === 0) {
    return null
  }

  const currentQuestion = questions.find(q => q.step === currentStep)
  const isFirstQuestion = currentStep === 1
  const isLastQuestion = currentStep === status.total_questions
  const canGoNext = currentQuestion && (
    currentQuestion.user_answer !== undefined || 
    currentQuestion.is_skipped
  )
  
  const isComplete = status.onboarding_completed
  
  return (
    <div className="flex flex-col gap-3 w-full p-4 bg-background border-t border-border">
      {/* Progress Indicator */}
      <div className="text-center">
        <div className="text-sm font-medium">
          Question {currentStep} of {status.total_questions}
        </div>
        <div className="text-xs text-muted-foreground">
          {status.completed_questions} answered • {status.skipped_questions} skipped
        </div>
      </div>

      {/* Navigation Buttons */}
      <div className="flex gap-2">
        {/* Previous Button */}
        {!isFirstQuestion && (
          <Button
            variant="outline"
            onClick={onPrevious}
            disabled={isSubmitting}
            className="flex-1"
          >
            <ChevronLeft className="h-4 w-4 mr-2" />
            Previous
          </Button>
        )}

        {/* Next/Complete Button */}
        {isComplete ? (
          <Button
            onClick={onComplete}
            disabled={isSubmitting}
            className={`flex-1 bg-green-600 hover:bg-green-700 ${!isFirstQuestion ? '' : 'ml-0'}`}
          >
            <CheckCircle className="h-4 w-4 mr-2" />
            Complete Setup
          </Button>
        ) : (
          <Button
            variant={canGoNext ? "default" : "outline"}
            onClick={onNext}
            disabled={isSubmitting || !canGoNext}
            className={`flex-1 ${!isFirstQuestion ? '' : 'ml-0'}`}
          >
            {isLastQuestion ? 'Finish' : 'Next'}
            {isLastQuestion ? (
              <CheckCircle className="h-4 w-4 ml-2" />
            ) : (
              <ChevronRight className="h-4 w-4 ml-2" />
            )}
          </Button>
        )}
      </div>
    </div>
  )
}