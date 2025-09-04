import { Building2, ArrowLeft } from 'lucide-react'
import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

export default function OrganizationNotFound() {
  return (
    <div className="flex h-screen items-center justify-center p-6">
      <Card className="w-full max-w-md text-center">
        <CardHeader className="pb-4">
          <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-destructive/10">
            <Building2 className="h-8 w-8 text-destructive" />
          </div>
          <CardTitle className="text-xl">Organization Not Found</CardTitle>
          <CardDescription className="text-base">
            The organization you're looking for doesn't exist or you don't have access to it.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-sm text-muted-foreground">
            This could happen if:
          </p>
          <ul className="text-sm text-muted-foreground space-y-1 text-left">
            <li>• The organization URL is incorrect</li>
            <li>• You don't have permission to access this organization</li>
            <li>• The organization has been deleted or archived</li>
          </ul>
          
          <div className="flex flex-col gap-2 pt-4">
            <Button asChild>
              <Link href="/">
                <ArrowLeft className="mr-2 h-4 w-4" />
                Back to Organizations
              </Link>
            </Button>
            <Button variant="outline" asChild>
              <Link href="/auth/signin">
                Sign In with Different Account
              </Link>
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}