import { FolderOpen, ArrowLeft, Building2 } from 'lucide-react'
import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

export default function ProjectNotFound() {
  return (
    <div className="flex h-screen items-center justify-center p-6">
      <Card className="w-full max-w-md text-center">
        <CardHeader className="pb-4">
          <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-destructive/10">
            <FolderOpen className="h-8 w-8 text-destructive" />
          </div>
          <CardTitle className="text-xl">Project Not Found</CardTitle>
          <CardDescription className="text-base">
            The project you're looking for doesn't exist or you don't have access to it.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-sm text-muted-foreground">
            This could happen if:
          </p>
          <ul className="text-sm text-muted-foreground space-y-1 text-left">
            <li>• The project URL is incorrect</li>
            <li>• You don't have permission to access this project</li>
            <li>• The project has been deleted or archived</li>
            <li>• The organization doesn't contain this project</li>
          </ul>
          
          <div className="flex flex-col gap-2 pt-4">
            <Button asChild>
              <Link href="../">
                <Building2 className="mr-2 h-4 w-4" />
                Back to Organization
              </Link>
            </Button>
            <Button variant="outline" asChild>
              <Link href="/">
                <ArrowLeft className="mr-2 h-4 w-4" />
                All Organizations
              </Link>
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}