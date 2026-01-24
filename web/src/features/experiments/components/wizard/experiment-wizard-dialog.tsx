'use client'

import { useEffect } from 'react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog'
import { ExperimentWizardProvider, useExperimentWizard } from '../../context/experiment-wizard-context'
import { WizardStepper } from './wizard-stepper'
import { WizardNavigation } from './wizard-navigation'
import { ConfigStep } from './step-config'
import { DatasetStep } from './step-dataset'
import { EvaluatorsStep } from './step-evaluators'
import { ReviewStep } from './step-review'

interface ExperimentWizardDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess?: () => void
}

export function ExperimentWizardDialog({
  open,
  onOpenChange,
  onSuccess,
}: ExperimentWizardDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-2xl max-h-[90vh] overflow-hidden flex flex-col">
        <ExperimentWizardProvider>
          <WizardContent onOpenChange={onOpenChange} onSuccess={onSuccess} />
        </ExperimentWizardProvider>
      </DialogContent>
    </Dialog>
  )
}

function WizardContent({
  onOpenChange,
  onSuccess,
}: {
  onOpenChange: (open: boolean) => void
  onSuccess?: () => void
}) {
  const { state, reset } = useExperimentWizard()

  // Reset wizard state when dialog opens
  useEffect(() => {
    reset()
  }, [reset])

  const handleSuccess = () => {
    onSuccess?.()
    onOpenChange(false)
    reset()
  }

  const renderStep = () => {
    switch (state.currentStep) {
      case 1:
        return <ConfigStep />
      case 2:
        return <DatasetStep />
      case 3:
        return <EvaluatorsStep />
      case 4:
        return <ReviewStep onSuccess={handleSuccess} />
      default:
        return <ConfigStep />
    }
  }

  const getStepTitle = () => {
    switch (state.currentStep) {
      case 1:
        return 'Configure Experiment'
      case 2:
        return 'Select Dataset'
      case 3:
        return 'Add Evaluators'
      case 4:
        return 'Review & Create'
      default:
        return 'Create Experiment'
    }
  }

  const getStepDescription = () => {
    switch (state.currentStep) {
      case 1:
        return 'Choose a prompt and configure the model settings.'
      case 2:
        return 'Select a dataset and map variables to prompt placeholders.'
      case 3:
        return 'Add evaluators to score the experiment outputs.'
      case 4:
        return 'Review your configuration and create the experiment.'
      default:
        return ''
    }
  }

  return (
    <>
      <DialogHeader>
        <DialogTitle>{getStepTitle()}</DialogTitle>
        <DialogDescription>{getStepDescription()}</DialogDescription>
      </DialogHeader>

      <WizardStepper />

      <div className="flex-1 overflow-y-auto py-4 min-h-0">
        {renderStep()}
      </div>

      <WizardNavigation />
    </>
  )
}
