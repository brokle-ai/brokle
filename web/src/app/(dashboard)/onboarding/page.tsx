'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { QuestionCard } from '@/components/onboarding'
import { useAuth } from '@/context/auth-context'
import { useOrganization } from '@/context/organization-context'
import { api } from '@/lib/api'
import type { OnboardingQuestion } from '@/types/onboarding'

export default function OnboardingPage() {
  const router = useRouter()
  const { user } = useAuth()
  const { organizations } = useOrganization()
  
  // Simplified onboarding state
  const [questions, setQuestions] = useState<OnboardingQuestion[]>([])
  const [currentStep, setCurrentStep] = useState(1)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  // Load onboarding data
  const loadOnboardingData = async () => {
    try {
      setLoading(true)
      setError(null)
      
      const questionsData = await api.onboarding.getQuestions()
      setQuestions(questionsData)
      
      // Set current step to next unanswered question or first step
      const nextQuestion = questionsData.find(q => !q.user_answer && !q.is_skipped)
      setCurrentStep(nextQuestion?.step || 1)
      
      // If all questions are answered/skipped, redirect to main app
      const allCompleted = questionsData.every(q => q.user_answer || q.is_skipped)
      if (allCompleted && questionsData.length > 0) {
        // Redirect to main app instead of showing organization creation
        if (organizations.length > 0) {
          router.push(`/${organizations[0].slug}`)
        } else {
          router.push('/')
        }
        return
      }
      
    } catch (err) {
      console.error('Error loading onboarding data:', err)
      setError('Failed to load onboarding questions. Please try again.')
    } finally {
      setLoading(false)
    }
  }

  // Submit response for a question
  const handleSubmitResponse = async (questionId: string, value: string | string[]) => {
    try {
      setIsSubmitting(true)
      await api.onboarding.submitResponse(questionId, value)
      
      // Refresh data to get updated progress
      await loadOnboardingData()
      
      // Auto-advance to next question linearly (but not on last question)
      if (currentStep < questions.length) {
        const nextStep = currentStep + 1
        setCurrentStep(nextStep)
      }
      // Note: Last question will be handled by handleFinish() after submit
    } catch (err) {
      console.error('Error submitting response:', err)
      setError('Failed to submit response. Please try again.')
    } finally {
      setIsSubmitting(false)
    }
  }

  // Skip a question
  const handleSkipQuestion = async (questionId: string) => {
    try {
      setIsSubmitting(true)
      await api.onboarding.skipQuestion(questionId)
      
      // Refresh data to get updated progress
      await loadOnboardingData()
      
      // Auto-advance to next question linearly (but not on last question)
      if (currentStep < questions.length) {
        const nextStep = currentStep + 1
        setCurrentStep(nextStep)
      }
      // Note: Last question will be handled by handleFinish() after submit
    } catch (err) {
      console.error('Error skipping question:', err)
      setError('Failed to skip question. Please try again.')
    } finally {
      setIsSubmitting(false)
    }
  }

  // Navigate to previous question
  const handlePrevious = () => {
    if (currentStep > 1) {
      setCurrentStep(currentStep - 1)
    }
  }

  // Check if we can go back
  const canGoBack = currentStep > 1

  // Handle finish - complete onboarding and redirect to main routing
  const handleFinish = () => {
    router.push('/') // Let main routing logic decide where to go
  }

  // Load data on mount
  useEffect(() => {
    loadOnboardingData()
  }, [])

  // Loading state
  if (loading) {
    return (
      <div className="min-h-screen bg-white flex items-center justify-center">
        <div className="text-center">
          <div className="text-gray-500">Loading...</div>
        </div>
      </div>
    )
  }

  // Error state
  if (error) {
    return (
      <div className="min-h-screen bg-white flex items-center justify-center">
        <div className="text-center">
          <div className="text-red-600 mb-4">{error}</div>
          <Button onClick={loadOnboardingData} variant="outline">
            Try Again
          </Button>
        </div>
      </div>
    )
  }

  // Get current question
  const currentQuestion = questions.find(q => q.step === currentStep)

  // If no current question, something went wrong or all completed
  if (!currentQuestion) {
    return (
      <div className="min-h-screen bg-white flex items-center justify-center">
        <div className="text-center">
          <div className="text-gray-500">No questions available</div>
        </div>
      </div>
    )
  }

  // Main onboarding flow
  return (
    <div className="min-h-screen bg-white">
      <div className="container mx-auto px-6 py-8">
        {/* Header */}
        <div className="text-center mb-8">
          <h1 className="text-2xl font-semibold text-gray-900 mb-4">
            Welcome, {user?.firstName || user?.name || 'there'}!
          </h1>
          <p className="text-gray-600 max-w-md mx-auto">
            Help us personalize your experience by answering a few quick questions. This will only take a minute.
          </p>
        </div>

        {/* Question */}
        <QuestionCard
          question={currentQuestion}
          onSubmitResponse={handleSubmitResponse}
          onSkipQuestion={handleSkipQuestion}
          onPrevious={handlePrevious}
          onFinish={handleFinish}
          isSubmitting={isSubmitting}
          canGoBack={canGoBack}
          currentStep={currentStep}
          totalQuestions={questions.length}
        />
      </div>
    </div>
  )
}