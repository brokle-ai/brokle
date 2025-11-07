'use client'

import { useState } from 'react'
import { Code, Plus, Trash2, Download, Upload, RotateCcw } from 'lucide-react'
import { useWorkspace } from '@/context/workspace-context'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { toast } from 'sonner'

interface Setting {
  key: string
  value: string
  created_at: string
  updated_at: string
}

const MOCK_SETTINGS: Setting[] = [
  {
    key: 'webhook_url',
    value: 'https://api.example.com/webhooks/org',
    created_at: '2024-03-01T10:00:00Z',
    updated_at: '2024-03-01T10:00:00Z'
  },
  {
    key: 'data_retention_days',
    value: '90',
    created_at: '2024-03-05T14:30:00Z',
    updated_at: '2024-03-10T09:15:00Z'
  },
]

export function OrganizationAdvancedSection() {
  const { currentOrganization } = useWorkspace()
  const [settings, setSettings] = useState<Setting[]>(MOCK_SETTINGS)
  const [isAddOpen, setIsAddOpen] = useState(false)
  const [newKey, setNewKey] = useState('')
  const [newValue, setNewValue] = useState('')

  if (!currentOrganization) {
    return null
  }

  const addSetting = () => {
    if (!newKey.trim() || !newValue.trim()) {
      toast.error('Please enter both key and value')
      return
    }

    if (settings.some(s => s.key === newKey)) {
      toast.error('A setting with this key already exists')
      return
    }

    const newSetting: Setting = {
      key: newKey,
      value: newValue,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    }

    setSettings([...settings, newSetting])
    setNewKey('')
    setNewValue('')
    setIsAddOpen(false)
    toast.success('Setting added successfully')
  }

  const deleteSetting = (key: string) => {
    setSettings(settings.filter(s => s.key !== key))
    toast.success('Setting deleted')
  }

  const exportSettings = () => {
    const data = JSON.stringify({ settings }, null, 2)
    const blob = new Blob([data], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${currentOrganization.name}-settings-${new Date().toISOString().split('T')[0]}.json`
    a.click()
    URL.revokeObjectURL(url)
    toast.success('Settings exported successfully')
  }

  return (
    <div className="space-y-8">
      {/* Key-Value Settings Editor */}
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-medium">Custom Settings</h3>
          <Dialog open={isAddOpen} onOpenChange={setIsAddOpen}>
            <DialogTrigger asChild>
              <Button>
                <Plus className="mr-2 h-4 w-4" />
                Add Setting
              </Button>
            </DialogTrigger>

            <DialogContent>
              <DialogHeader>
                <DialogTitle>Add Custom Setting</DialogTitle>
                <DialogDescription>
                  Create a new key-value setting for your organization
                </DialogDescription>
              </DialogHeader>

              <div className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="settingKey">Setting Key *</Label>
                  <Input
                    id="settingKey"
                    value={newKey}
                    onChange={(e) => setNewKey(e.target.value)}
                    placeholder="e.g., webhook_url, retention_days"
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="settingValue">Setting Value *</Label>
                  <Textarea
                    id="settingValue"
                    value={newValue}
                    onChange={(e) => setNewValue(e.target.value)}
                    placeholder="Enter setting value"
                    rows={3}
                  />
                </div>
              </div>

              <DialogFooter>
                <Button variant="outline" onClick={() => setIsAddOpen(false)}>
                  Cancel
                </Button>
                <Button onClick={addSetting}>Add Setting</Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </div>

        {settings.length > 0 ? (
          <div className="rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Key</TableHead>
                  <TableHead>Value</TableHead>
                  <TableHead>Last Updated</TableHead>
                  <TableHead className="w-[100px]"></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {settings.map((setting) => (
                  <TableRow key={setting.key}>
                    <TableCell className="font-mono text-sm">{setting.key}</TableCell>
                    <TableCell className="max-w-xs truncate">{setting.value}</TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {new Date(setting.updated_at).toLocaleDateString()}
                    </TableCell>
                    <TableCell>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => deleteSetting(setting.key)}
                        className="text-red-600 hover:text-red-700"
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        ) : (
          <div className="text-center py-8 rounded-lg border text-muted-foreground">
            <Code className="mx-auto h-8 w-8 mb-2 opacity-50" />
            <div className="text-sm">No custom settings configured</div>
            <div className="text-xs">Click "Add Setting" to create your first custom setting</div>
          </div>
        )}
      </div>

      {/* Data Management */}
      <div className="space-y-4">
        <h3 className="text-lg font-medium">Data Management</h3>

        <div className="flex flex-wrap gap-3">
          <Button variant="outline" onClick={exportSettings}>
            <Download className="mr-2 h-4 w-4" />
            Export Settings
          </Button>

          <Button variant="outline" disabled>
            <Upload className="mr-2 h-4 w-4" />
            Import Settings
          </Button>

          <Button variant="outline" disabled>
            <RotateCcw className="mr-2 h-4 w-4" />
            Reset to Defaults
          </Button>
        </div>
      </div>

      {/* Webhook Configuration */}
      <div className="space-y-4">
        <h3 className="text-lg font-medium">Webhooks</h3>

        <div className="space-y-2">
          <Label htmlFor="orgWebhook">Organization Webhook URL</Label>
          <Input
            id="orgWebhook"
            type="url"
            placeholder="https://your-app.com/webhooks/brokle/org"
          />
          <p className="text-xs text-muted-foreground">
            Receive notifications about organization-level events (member changes, plan updates, etc.)
          </p>
        </div>

        <Button variant="outline">Save Webhook Configuration</Button>
      </div>
    </div>
  )
}
