'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/context/auth-context'
import { useOnboardingForm } from '@/features/onboarding/hooks/useOnboardingForm'
import { QuestionCard } from '@/components/onboarding/question-card'
import { PageLoader } from '@/components/shared/loading'

export default function OnboardingPage() {
  const router = useRouter()
  const { user, isLoading: authLoading } = useAuth()
  const {
    state,
    currentQuestion,
    isFirstStep,
    effectiveTotalSteps,
    goBack,
    handleSubmitResponse,
    handleSkipQuestion,
    handleFinish,
  } = useOnboardingForm()

  // Guard: Redirect if already completed onboarding
  useEffect(() => {
    if (authLoading) return

    if (user?.onboardingCompletedAt) {
      router.push('/')
    }
  }, [user, authLoading, router])

  if (authLoading || (user && user.onboardingCompletedAt)) {
    return <PageLoader message="Loading..." />
  }

  if (!currentQuestion) {
    return <PageLoader message="Loading onboarding questions..." type="spinner" />
  }

  return (
    <div className="min-h-screen bg-background">
      <div className="container max-w-2xl mx-auto px-6 py-12">
        {/* Header */}
        <div className="text-center mb-12">
          <h1 className="text-3xl font-bold text-foreground mb-3">
            Welcome to Brokle, {user?.firstName || user?.name || 'there'}! 👋
          </h1>
          <p className="text-muted-foreground text-lg max-w-md mx-auto">
            Help us personalize your experience. This will only take a minute.
          </p>
        </div>

        {/* Question Card - handles everything (submit, skip, back buttons) */}
        <QuestionCard
          question={currentQuestion}
          onSubmitResponse={handleSubmitResponse}
          onSkipQuestion={handleSkipQuestion}
          onPrevious={goBack}
          onFinish={handleFinish}
          isSubmitting={state.isSubmitting}
          canGoBack={!isFirstStep}
          currentStep={state.currentStep + 1}
          totalQuestions={effectiveTotalSteps}
        />
      </div>
    </div>
  )
}
