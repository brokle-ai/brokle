'use client'

import { Progress } from '@/components/ui/progress'
import { Badge } from '@/components/ui/badge'
import { CheckCircle, Circle, Minus } from 'lucide-react'
import type { OnboardingStatus, OnboardingQuestion } from '@/types/onboarding'

interface OnboardingProgressProps {
  status: OnboardingStatus | null
  questions: OnboardingQuestion[]
  currentStep: number
}

export function OnboardingProgress({ status, questions, currentStep }: OnboardingProgressProps) {
  if (!status) {
    return (
      <div className="w-full space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-semibold">Loading Progress...</h3>
        </div>
        <Progress value={0} className="h-2" />
      </div>
    )
  }

  const completionPercentage = status.total_questions > 0 
    ? Math.round(((status.completed_questions + status.skipped_questions) / status.total_questions) * 100)
    : 0

  const getQuestionStatus = (question: OnboardingQuestion) => {
    if (question.user_answer !== undefined && question.user_answer !== null) {
      return 'completed'
    }
    if (question.is_skipped) {
      return 'skipped'
    }
    if (question.step === currentStep) {
      return 'current'
    }
    if (question.step < currentStep) {
      return 'pending'
    }
    return 'upcoming'
  }

  const getStatusIcon = (questionStatus: string) => {
    switch (questionStatus) {
      case 'completed':
        return <CheckCircle className="h-4 w-4 text-green-600" />
      case 'skipped':
        return <Minus className="h-4 w-4 text-gray-400" />
      case 'current':
        return <Circle className="h-4 w-4 text-blue-600 fill-blue-100" />
      default:
        return <Circle className="h-4 w-4 text-gray-300" />
    }
  }

  const getStatusColor = (questionStatus: string) => {
    switch (questionStatus) {
      case 'completed':
        return 'bg-green-100 text-green-800 border-green-200'
      case 'skipped':
        return 'bg-gray-100 text-gray-600 border-gray-200'
      case 'current':
        return 'bg-blue-100 text-blue-800 border-blue-200'
      default:
        return 'bg-gray-50 text-gray-500 border-gray-100'
    }
  }

  return (
    <div className="w-full space-y-6">
      {/* Overall Progress */}
      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-semibold">Onboarding Progress</h3>
          <Badge variant="outline" className="text-sm">
            {completionPercentage}% Complete
          </Badge>
        </div>
        
        <Progress value={completionPercentage} className="h-3" />
        
        <div className="flex justify-between text-sm text-muted-foreground">
          <span>Step {Math.min(currentStep, status.total_questions)} of {status.total_questions}</span>
          <span>
            {status.completed_questions} answered, {status.skipped_questions} skipped
          </span>
        </div>
      </div>

      {/* Step Indicators */}
      {questions.length > 0 && (
        <div className="space-y-3">
          <h4 className="font-medium text-sm text-muted-foreground">Questions</h4>
          <div className="grid gap-2">
            {questions.map((question) => {
              const questionStatus = getQuestionStatus(question)
              return (
                <div
                  key={question.id}
                  className={`flex items-center gap-3 p-3 rounded-lg border transition-colors ${getStatusColor(questionStatus)}`}
                >
                  {getStatusIcon(questionStatus)}
                  <div className="flex-1 min-w-0">
                    <p className="font-medium text-sm truncate">
                      {question.title}
                    </p>
                    <div className="flex items-center gap-2 mt-1">
                      <span className="text-xs opacity-75">
                        Step {question.step}
                      </span>
                      {!question.is_required && (
                        <Badge variant="secondary" className="text-xs h-4">
                          Optional
                        </Badge>
                      )}
                    </div>
                  </div>
                  {questionStatus === 'current' && (
                    <Badge variant="default" className="text-xs">
                      Current
                    </Badge>
                  )}
                </div>
              )
            })}
          </div>
        </div>
      )}

      {/* Summary Stats */}
      <div className="grid grid-cols-3 gap-4 p-4 bg-muted/30 rounded-lg">
        <div className="text-center">
          <div className="text-2xl font-bold text-green-600">
            {status.completed_questions}
          </div>
          <div className="text-xs text-muted-foreground">Answered</div>
        </div>
        <div className="text-center">
          <div className="text-2xl font-bold text-gray-500">
            {status.skipped_questions}
          </div>
          <div className="text-xs text-muted-foreground">Skipped</div>
        </div>
        <div className="text-center">
          <div className="text-2xl font-bold text-blue-600">
            {status.remaining_questions}
          </div>
          <div className="text-xs text-muted-foreground">Remaining</div>
        </div>
      </div>

      {/* Completion Message */}
      {status.onboarding_completed && (
        <div className="bg-green-50 border border-green-200 rounded-lg p-4 text-center">
          <CheckCircle className="h-8 w-8 text-green-600 mx-auto mb-2" />
          <p className="font-semibold text-green-800">Onboarding Complete!</p>
          <p className="text-green-600 text-sm mt-1">
            Thank you for taking the time to tell us about yourself.
          </p>
        </div>
      )}
    </div>
  )
}