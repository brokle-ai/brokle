'use client'

import { useAuth } from '@/context/auth-context'
import { useOnboardingForm } from '@/features/onboarding/hooks/useOnboardingForm'
import { QuestionCard } from '@/components/onboarding/question-card'

export default function OnboardingPage() {
  const { user } = useAuth()
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

  // Loading state
  if (!currentQuestion) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="text-center">
          <div className="text-muted-foreground">Loading onboarding...</div>
        </div>
      </div>
    )
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
