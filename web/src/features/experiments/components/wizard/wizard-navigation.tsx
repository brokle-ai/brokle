'use client'

import { Button } from '@/components/ui/button'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { useExperimentWizard } from '../../context/experiment-wizard-context'

export function WizardNavigation() {
  const { state, attemptNextStep, prevStep } = useExperimentWizard()

  const isFirstStep = state.currentStep === 1
  const isLastStep = state.currentStep === 4

  // Don't show navigation on review step (it has its own submit button)
  if (isLastStep) {
    return (
      <div className="flex justify-between pt-4 border-t">
        <Button variant="outline" onClick={prevStep}>
          <ChevronLeft className="mr-2 h-4 w-4" />
          Back
        </Button>
        <div />
      </div>
    )
  }

  return (
    <div className="flex justify-between pt-4 border-t">
      <Button
        variant="outline"
        onClick={prevStep}
        disabled={isFirstStep}
      >
        <ChevronLeft className="mr-2 h-4 w-4" />
        Back
      </Button>

      <Button onClick={attemptNextStep}>
        Next
        <ChevronRight className="ml-2 h-4 w-4" />
      </Button>
    </div>
  )
}
