import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'
import { Check } from 'lucide-react'
import { cn } from '@/lib/utils'

interface WizardProgressProps {
  currentStep: number
  steps: {
    number: number
    title: string
  }[]
}

export function WizardProgress({ currentStep, steps }: WizardProgressProps) {
  return (
    <Breadcrumb className="mb-6">
      <BreadcrumbList>
        {steps.map((step, index) => (
          <div key={step.number} className="flex items-center">
            <BreadcrumbItem>
              <BreadcrumbPage
                className={cn(
                  'flex items-center gap-2',
                  currentStep !== step.number
                    ? 'text-muted-foreground'
                    : 'font-semibold text-foreground'
                )}
              >
                {step.number}. {step.title}
                {currentStep > step.number && (
                  <Check className="h-4 w-4 text-green-600" />
                )}
              </BreadcrumbPage>
            </BreadcrumbItem>
            {index < steps.length - 1 && <BreadcrumbSeparator />}
          </div>
        ))}
      </BreadcrumbList>
    </Breadcrumb>
  )
}
