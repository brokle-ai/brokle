'use client'

import { Suspense, useState, useEffect } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { useAuth } from '@/hooks/auth/use-auth'
import { useWorkspace } from '@/context/workspace-context'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { WizardProgress } from '@/components/wizard/wizard-progress'
import { CreateOrgForm } from '@/components/wizard/create-org-form'
import { CreateProjectForm } from '@/components/wizard/create-project-form'
import { InviteMemberModal } from '@/components/organization/invite-member-modal'
import { PageLoader } from '@/components/shared/loading'
import { ChevronRight, Loader2 } from 'lucide-react'

function SetupWizardContent() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const { user } = useAuth()
  const { organizations, isLoading: orgsLoading } = useWorkspace()

  const [orgId, setOrgId] = useState<string | null>(null)

  // If user already has organizations, skip wizard
  useEffect(() => {
    if (orgsLoading) return

    if (organizations && organizations.length > 0) {
      // User already has orgs, go to dashboard
      router.push('/')
    }
  }, [organizations, orgsLoading, router])

  // Determine current step from URL state
  const step = searchParams.get('step')
  const currentStep = !orgId ? 1 : step === 'project' ? 3 : 2

  const wizardSteps = [
    { number: 1, title: 'Create Organization' },
    { number: 2, title: 'Invite Team Members' },
    { number: 3, title: 'Create Project' },
  ]

  if (!user || orgsLoading) {
    return <PageLoader message="Loading..." />
  }

  // If user has orgs, will redirect (show loader)
  if (organizations && organizations.length > 0) {
    return <PageLoader message="Redirecting..." />
  }

  return (
    <div className="min-h-screen bg-background">
      <div className="container max-w-4xl mx-auto px-6 py-12">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold mb-2">Setup Your Workspace</h1>
          <p className="text-muted-foreground">
            Let's get you started with your organization and first project
          </p>
        </div>

        {/* Progress */}
        <WizardProgress currentStep={currentStep} steps={wizardSteps} />

        {/* Step Content */}
        <Card>
          <CardContent className="p-6">
            {/* Step 1: Create Organization */}
            {currentStep === 1 && (
              <div className="space-y-6">
                <div>
                  <h2 className="text-2xl font-semibold mb-2">Create Your Organization</h2>
                  <p className="text-muted-foreground">
                    Organizations help you manage projects and team members
                  </p>
                </div>

                <CreateOrgForm
                  onSuccess={(createdOrgId) => {
                    setOrgId(createdOrgId)
                    router.push('/setup/wizard?step=invite')
                  }}
                />
              </div>
            )}

            {/* Step 2: Invite Members */}
            {currentStep === 2 && orgId && (
              <div className="space-y-6">
                <div>
                  <h2 className="text-2xl font-semibold mb-2">Invite Your Team</h2>
                  <p className="text-muted-foreground mb-4">
                    Collaborate with your team members (you can skip this for now)
                  </p>
                </div>

                <InviteMemberModal
                  trigger={
                    <Button variant="outline" className="w-full">
                      Invite Team Members
                    </Button>
                  }
                />

                <div className="flex justify-between pt-4 border-t">
                  <Button variant="ghost" onClick={() => router.back()}>
                    Back
                  </Button>
                  <div className="flex gap-2">
                    <Button variant="outline" onClick={() => router.push('/setup/wizard?step=project')}>
                      Skip for Now
                    </Button>
                    <Button onClick={() => router.push('/setup/wizard?step=project')}>
                      Continue <ChevronRight className="ml-2 h-4 w-4" />
                    </Button>
                  </div>
                </div>
              </div>
            )}

            {/* Step 3: Create Project */}
            {currentStep === 3 && orgId && (
              <div className="space-y-6">
                <div>
                  <h2 className="text-2xl font-semibold mb-2">Create Your First Project</h2>
                  <p className="text-muted-foreground mb-4">
                    Projects help you organize your AI applications and environments
                  </p>
                </div>

                <CreateProjectForm
                  organizationId={orgId}
                  onSuccess={(createdProjectId) => {
                    router.push('/')
                  }}
                />

                <div className="flex justify-start pt-4 border-t">
                  <Button variant="ghost" onClick={() => router.back()}>
                    Back
                  </Button>
                </div>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Help Text */}
        <p className="text-center text-sm text-muted-foreground mt-6">
          Need help? <a href="/docs" className="text-primary hover:underline">Check our documentation</a>
        </p>
      </div>
    </div>
  )
}

export default function SetupWizardPage() {
  return (
    <Suspense
      fallback={
        <div className="flex min-h-screen items-center justify-center">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      }
    >
      <SetupWizardContent />
    </Suspense>
  )
}
