'use client'

import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { Checkbox } from '@/components/ui/checkbox'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { ArrowRight, ArrowLeft } from 'lucide-react'
import type { OnboardingQuestion } from '@/types/onboarding'

interface QuestionCardProps {
  question: OnboardingQuestion
  onSubmitResponse: (questionId: string, value: string | string[]) => Promise<void>
  onSkipQuestion: (questionId: string) => Promise<void>
  onPrevious?: () => void
  onFinish?: () => void
  isSubmitting?: boolean
  canGoBack?: boolean
  currentStep?: number
  totalQuestions?: number
}

export function QuestionCard({ 
  question, 
  onSubmitResponse, 
  onSkipQuestion,
  onPrevious,
  onFinish,
  isSubmitting = false,
  canGoBack = false,
  currentStep = 1,
  totalQuestions = 1
}: QuestionCardProps) {
  const [singleChoiceValue, setSingleChoiceValue] = useState<string>(
    question.user_answer as string || ''
  )
  const [multipleChoiceValues, setMultipleChoiceValues] = useState<string[]>(() => {
    if (!question.user_answer) return []
    if (Array.isArray(question.user_answer)) return question.user_answer
    if (typeof question.user_answer === 'string') {
      try {
        const parsed = JSON.parse(question.user_answer)
        return Array.isArray(parsed) ? parsed : []
      } catch {
        return []
      }
    }
    return []
  })
  const [textValue, setTextValue] = useState<string>(
    question.user_answer as string || ''
  )

  // Sync state with question prop changes for navigation between questions
  useEffect(() => {
    if (question.is_skipped) {
      setSingleChoiceValue('')
    } else {
      setSingleChoiceValue(question.user_answer as string || '')
    }
  }, [question.user_answer, question.id, question.is_skipped])

  useEffect(() => {
    if (question.is_skipped) {
      setMultipleChoiceValues([])
    } else if (question.user_answer) {
      if (Array.isArray(question.user_answer)) {
        setMultipleChoiceValues(question.user_answer)
      } else if (typeof question.user_answer === 'string') {
        try {
          const parsed = JSON.parse(question.user_answer)
          setMultipleChoiceValues(Array.isArray(parsed) ? parsed : [])
        } catch {
          setMultipleChoiceValues([])
        }
      } else {
        setMultipleChoiceValues([])
      }
    } else {
      setMultipleChoiceValues([])
    }
  }, [question.user_answer, question.id, question.is_skipped])

  useEffect(() => {
    if (question.is_skipped) {
      setTextValue('')
    } else {
      setTextValue(question.user_answer as string || '')
    }
  }, [question.user_answer, question.id, question.is_skipped])

  const handleSingleChoiceChange = (value: string) => {
    setSingleChoiceValue(value)
  }

  const handleSingleChoiceSubmit = async () => {
    if (!singleChoiceValue) return
    await onSubmitResponse(question.id, singleChoiceValue)
  }

  const handleMultipleChoiceChange = (option: string) => {
    const newValues = multipleChoiceValues.includes(option)
      ? multipleChoiceValues.filter(v => v !== option)
      : [...multipleChoiceValues, option]
    setMultipleChoiceValues(newValues)
  }

  const handleMultipleChoiceSubmit = async () => {
    await onSubmitResponse(question.id, multipleChoiceValues)
  }

  const handleTextSubmit = async () => {
    await onSubmitResponse(question.id, textValue)
  }

  const handleSkip = async () => {
    await onSkipQuestion(question.id)
  }

  // Check if this is the last question
  const isLastQuestion = currentStep === totalQuestions

  // Handle finish action
  const handleFinish = () => {
    if (onFinish) {
      onFinish()
    }
  }

  // Unified submit handler for all question types
  const handleSubmit = async () => {
    if (question.question_type === 'single_choice') {
      await handleSingleChoiceSubmit()
    } else if (question.question_type === 'multiple_choice') {
      await handleMultipleChoiceSubmit()
    } else if (question.question_type === 'text') {
      await handleTextSubmit()
    }
  }

  // Combined submit and finish for last question
  const handleFinishWithSubmit = async () => {
    await handleSubmit()
    handleFinish()
  }

  // Check if current answer can be submitted
  const canSubmit = () => {
    // Special case: Last question that's optional - always allow finish
    if (isLastQuestion && !question.is_required) {
      return true
    }

    // Simple logic: Enable if has valid answer (regardless of source)
    let hasAnswer = false
    if (question.question_type === 'single_choice') {
      hasAnswer = !!singleChoiceValue
    } else if (question.question_type === 'multiple_choice') {
      hasAnswer = multipleChoiceValues.length > 0
    } else if (question.question_type === 'text') {
      hasAnswer = !!textValue.trim()
    }

    return hasAnswer
  }


  return (
    <div className="w-full max-w-lg mx-auto">
      <Card className="min-h-[400px] border border-gray-200 shadow-sm transition-all duration-200">
        <CardContent className="p-6 min-h-[400px] flex flex-col">
          <div className="flex-1 space-y-6">
            {/* Progress Indicator */}
            <div className="text-center">
              <div className="text-sm text-muted-foreground mb-2">
                Question {currentStep} of {totalQuestions}
              </div>
              <h2 className="text-xl font-medium text-gray-900">
                {question.title}
              </h2>
              {question.description && (
                <p className="text-sm text-muted-foreground mt-2">
                  {question.description}
                </p>
              )}
            </div>
            
            {/* Single Choice Questions */}
            {question.question_type === 'single_choice' && question.options && (
              <RadioGroup
                value={singleChoiceValue}
                onValueChange={handleSingleChoiceChange}
                disabled={isSubmitting}
                className="space-y-3"
              >
                {question.options.map((option) => (
                  <div key={option} className="flex items-center space-x-3 text-left">
                    <RadioGroupItem value={option} id={option} />
                    <Label
                      htmlFor={option}
                      className="text-base text-gray-700 cursor-pointer flex-1"
                    >
                      {option}
                    </Label>
                  </div>
                ))}
              </RadioGroup>
            )}

            {/* Multiple Choice Questions */}
            {question.question_type === 'multiple_choice' && question.options && (
              <div className="space-y-3">
                {question.options.map((option) => (
                  <div key={option} className="flex items-center space-x-3 text-left">
                    <Checkbox
                      id={option}
                      checked={multipleChoiceValues.includes(option)}
                      onCheckedChange={() => handleMultipleChoiceChange(option)}
                      disabled={isSubmitting}
                    />
                    <Label
                      htmlFor={option}
                      className="text-base text-gray-700 cursor-pointer flex-1"
                    >
                      {option}
                    </Label>
                  </div>
                ))}
              </div>
            )}

            {/* Text Questions */}
            {question.question_type === 'text' && (
              <Textarea
                placeholder="Enter your answer here..."
                value={textValue}
                onChange={(e) => setTextValue(e.target.value)}
                disabled={isSubmitting}
                className="min-h-[120px]"
              />
            )}

            {/* Unified Navigation Footer */}
            <div className="flex items-center justify-between pt-6 mt-6 border-t">
              {/* Left: Back */}
              <Button
                variant="ghost"
                onClick={onPrevious}
                disabled={!canGoBack || isSubmitting}
                className="gap-2"
              >
                <ArrowLeft className="h-4 w-4" />
                Back
              </Button>

              {/* Right: Skip + Continue */}
              <div className="flex items-center gap-2">
                {!question.is_required && !isLastQuestion && (
                  <Button
                    variant="ghost"
                    onClick={handleSkip}
                    disabled={isSubmitting}
                  >
                    Skip
                  </Button>
                )}
                <Button
                  onClick={isLastQuestion ? handleFinishWithSubmit : handleSubmit}
                  disabled={isSubmitting || !canSubmit()}
                  className="gap-2"
                >
                  {isLastQuestion ? 'Finish' : 'Continue'}
                  <ArrowRight className="h-4 w-4" />
                </Button>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}