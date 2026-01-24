'use client'

import { Check, Circle } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useExperimentWizard } from '../../context/experiment-wizard-context'

const STEPS = [
  { number: 1, label: 'Configuration' },
  { number: 2, label: 'Dataset' },
  { number: 3, label: 'Evaluators' },
  { number: 4, label: 'Review' },
] as const

export function WizardStepper() {
  const { state, goToStep, isStepComplete, validationState } = useExperimentWizard()

  const getStepStatus = (stepNumber: number): 'complete' | 'current' | 'upcoming' | 'invalid' => {
    if (stepNumber < state.currentStep && isStepComplete(stepNumber)) {
      return 'complete'
    }
    if (stepNumber === state.currentStep) {
      return 'current'
    }
    // Check if step has validation errors (for visited steps)
    if (state.completedSteps.has(stepNumber)) {
      const validation =
        stepNumber === 1 ? validationState.step1 :
        stepNumber === 2 ? validationState.step2 :
        stepNumber === 3 ? validationState.step3 : null
      if (validation && !validation.isValid) {
        return 'invalid'
      }
      return 'complete'
    }
    return 'upcoming'
  }

  const canNavigateTo = (stepNumber: number): boolean => {
    // Can always go back
    if (stepNumber < state.currentStep) return true
    // Can only go forward if current step is complete
    if (stepNumber === state.currentStep + 1 && isStepComplete(state.currentStep)) return true
    // Can go to any completed step
    if (state.completedSteps.has(stepNumber)) return true
    return false
  }

  return (
    <nav aria-label="Progress" className="px-4">
      <ol className="flex items-center justify-between">
        {STEPS.map((step, index) => {
          const status = getStepStatus(step.number)
          const isClickable = canNavigateTo(step.number)

          return (
            <li key={step.number} className="flex items-center">
              <button
                type="button"
                onClick={() => isClickable && goToStep(step.number as 1 | 2 | 3 | 4)}
                disabled={!isClickable}
                className={cn(
                  'group flex flex-col items-center gap-2',
                  isClickable && 'cursor-pointer',
                  !isClickable && 'cursor-not-allowed opacity-60'
                )}
              >
                <span
                  className={cn(
                    'flex h-10 w-10 items-center justify-center rounded-full border-2 text-sm font-medium transition-all',
                    status === 'complete' && 'border-green-500 bg-green-500 text-white',
                    status === 'current' && 'border-primary bg-primary text-primary-foreground',
                    status === 'invalid' && 'border-yellow-500 bg-yellow-50 text-yellow-700',
                    status === 'upcoming' && 'border-muted-foreground/30 bg-background text-muted-foreground',
                    isClickable && status === 'upcoming' && 'group-hover:border-primary/50'
                  )}
                >
                  {status === 'complete' ? (
                    <Check className="h-5 w-5" />
                  ) : (
                    <span>{step.number}</span>
                  )}
                </span>
                <span
                  className={cn(
                    'text-sm font-medium',
                    status === 'current' && 'text-primary',
                    status === 'complete' && 'text-green-600',
                    status === 'invalid' && 'text-yellow-600',
                    status === 'upcoming' && 'text-muted-foreground'
                  )}
                >
                  {step.label}
                </span>
              </button>

              {/* Connector line */}
              {index < STEPS.length - 1 && (
                <div
                  className={cn(
                    'mx-4 h-0.5 flex-1 min-w-[60px]',
                    status === 'complete' ? 'bg-green-500' : 'bg-muted-foreground/30'
                  )}
                />
              )}
            </li>
          )
        })}
      </ol>
    </nav>
  )
}
