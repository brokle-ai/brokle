'use client'

import { useState } from 'react'
import { 
  Download, 
  Upload, 
  FileText, 
  Database, 
  Settings, 
  Users, 
  BarChart3, 
  CheckCircle, 
  AlertCircle,
  RefreshCw,
  File,
  Folder
} from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import { Checkbox } from '@/components/ui/checkbox'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Separator } from '@/components/ui/separator'
import { toast } from 'sonner'
import { cn } from '@/lib/utils'

interface ExportItem {
  id: string
  name: string
  description: string
  icon: React.ComponentType<{ className?: string }>
  size: string
  included: boolean
  required?: boolean
}

interface ImportResult {
  success: boolean
  imported: number
  skipped: number
  errors: string[]
}

const EXPORT_ITEMS: ExportItem[] = [
  {
    id: 'projects',
    name: 'Projects',
    description: 'All project configurations, settings, and metadata',
    icon: Folder,
    size: '2.4 MB',
    included: true,
    required: true
  },
  {
    id: 'analytics',
    name: 'Analytics Data',
    description: 'Usage metrics, performance data, and statistics',
    icon: BarChart3,
    size: '15.7 MB',
    included: true
  },
  {
    id: 'settings',
    name: 'Organization Settings',
    description: 'Billing info, preferences, and configurations',
    icon: Settings,
    size: '0.1 MB',
    included: true
  },
  {
    id: 'members',
    name: 'Members & Permissions',
    description: 'User roles, permissions, and access levels',
    icon: Users,
    size: '0.3 MB',
    included: true
  },
  {
    id: 'logs',
    name: 'Request Logs',
    description: 'API request history and audit trails (last 30 days)',
    icon: FileText,
    size: '45.2 MB',
    included: false
  },
  {
    id: 'cache',
    name: 'Cache Data',
    description: 'Semantic cache entries and embeddings',
    icon: Database,
    size: '8.9 MB',
    included: false
  }
]

interface ExportImportManagerProps {
  organizationId?: string
  projectId?: string
  level: 'organization' | 'project'
}

export function ExportImportManager({ organizationId, projectId, level }: ExportImportManagerProps) {
  const [exportItems, setExportItems] = useState<ExportItem[]>(EXPORT_ITEMS)
  const [exportFormat, setExportFormat] = useState<'json' | 'csv' | 'excel'>('json')
  const [isExporting, setIsExporting] = useState(false)
  const [exportProgress, setExportProgress] = useState(0)
  const [selectedFile, setSelectedFile] = useState<File | null>(null)
  const [isImporting, setIsImporting] = useState(false)
  const [importProgress, setImportProgress] = useState(0)
  const [importResult, setImportResult] = useState<ImportResult | null>(null)
  const [isPreviewOpen, setIsPreviewOpen] = useState(false)

  const toggleExportItem = (itemId: string) => {
    setExportItems(items =>
      items.map(item =>
        item.id === itemId && !item.required
          ? { ...item, included: !item.included }
          : item
      )
    )
  }

  const handleExport = async () => {
    const includedItems = exportItems.filter(item => item.included)
    if (includedItems.length === 0) {
      toast.error('Please select at least one item to export')
      return
    }

    setIsExporting(true)
    setExportProgress(0)

    try {
      // Simulate export progress
      const steps = includedItems.length
      for (let i = 0; i < steps; i++) {
        await new Promise(resolve => setTimeout(resolve, 1000))
        setExportProgress(((i + 1) / steps) * 100)
      }

      // Create export data
      const exportData = {
        metadata: {
          exported_at: new Date().toISOString(),
          level,
          organization_id: organizationId,
          project_id: projectId,
          format: exportFormat,
          version: '1.0'
        },
        data: includedItems.reduce((acc, item) => {
          acc[item.id] = {
            name: item.name,
            description: item.description,
            size: item.size,
            data: `Mock ${item.name} data would be here...`
          }
          return acc
        }, {} as Record<string, any>)
      }

      // Download file
      let blob: Blob
      let filename: string
      let mimeType: string

      switch (exportFormat) {
        case 'json':
          blob = new Blob([JSON.stringify(exportData, null, 2)], { type: 'application/json' })
          filename = `${level}-export-${new Date().toISOString().split('T')[0]}.json`
          mimeType = 'application/json'
          break
        case 'csv':
          const csvData = 'Name,Description,Size,Data\n' + 
            includedItems.map(item => 
              `"${item.name}","${item.description}","${item.size}","Mock data"`
            ).join('\n')
          blob = new Blob([csvData], { type: 'text/csv' })
          filename = `${level}-export-${new Date().toISOString().split('T')[0]}.csv`
          mimeType = 'text/csv'
          break
        case 'excel':
          // For demo purposes, we'll export as JSON with .xlsx extension
          blob = new Blob([JSON.stringify(exportData, null, 2)], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' })
          filename = `${level}-export-${new Date().toISOString().split('T')[0]}.xlsx`
          mimeType = 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
          break
        default:
          throw new Error('Unsupported export format')
      }

      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = filename
      a.click()
      URL.revokeObjectURL(url)

      toast.success(`Export completed! Downloaded ${includedItems.length} data types.`)

    } catch (error) {
      console.error('Export failed:', error)
      toast.error('Export failed. Please try again.')
    } finally {
      setIsExporting(false)
      setExportProgress(0)
    }
  }

  const handleFileSelect = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (file) {
      if (!file.name.endsWith('.json') && !file.name.endsWith('.csv') && !file.name.endsWith('.xlsx')) {
        toast.error('Please select a valid export file (.json, .csv, or .xlsx)')
        return
      }
      setSelectedFile(file)
      setImportResult(null)
    }
  }

  const handleImport = async () => {
    if (!selectedFile) {
      toast.error('Please select a file to import')
      return
    }

    setIsImporting(true)
    setImportProgress(0)
    setImportResult(null)

    try {
      // Simulate import progress
      const steps = 5
      for (let i = 0; i < steps; i++) {
        await new Promise(resolve => setTimeout(resolve, 800))
        setImportProgress(((i + 1) / steps) * 100)
      }

      // Simulate import results
      const mockResult: ImportResult = {
        success: true,
        imported: Math.floor(Math.random() * 50) + 20,
        skipped: Math.floor(Math.random() * 5),
        errors: Math.random() > 0.7 ? [
          'Unable to import cached embeddings: format mismatch',
          '3 member permissions could not be restored: users no longer exist'
        ] : []
      }

      setImportResult(mockResult)
      
      if (mockResult.success) {
        toast.success(`Import completed! ${mockResult.imported} items imported successfully.`)
      } else {
        toast.error('Import completed with errors. Check the results below.')
      }

    } catch (error) {
      console.error('Import failed:', error)
      toast.error('Import failed. Please check your file and try again.')
      setImportResult({
        success: false,
        imported: 0,
        skipped: 0,
        errors: ['File format not recognized or corrupted']
      })
    } finally {
      setIsImporting(false)
      setImportProgress(0)
    }
  }

  const previewImport = () => {
    if (!selectedFile) return
    setIsPreviewOpen(true)
  }

  const totalExportSize = exportItems
    .filter(item => item.included)
    .reduce((total, item) => {
      const size = parseFloat(item.size)
      return total + size
    }, 0)

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between">
        <div>
          <h2 className="text-2xl font-bold text-foreground">Data Export & Import</h2>
          <p className="text-muted-foreground">
            Export your {level} data or import from backup files
          </p>
        </div>
      </div>

      <Tabs defaultValue="export" className="space-y-6">
        <TabsList>
          <TabsTrigger value="export">Export Data</TabsTrigger>
          <TabsTrigger value="import">Import Data</TabsTrigger>
        </TabsList>

        <TabsContent value="export" className="space-y-6">
          {/* Export Configuration */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Download className="h-5 w-5" />
                Export Configuration
              </CardTitle>
              <CardDescription>
                Select the data you want to export and choose your preferred format
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              {/* Format Selection */}
              <div className="space-y-2">
                <Label>Export Format</Label>
                <Select value={exportFormat} onValueChange={(value: any) => setExportFormat(value)}>
                  <SelectTrigger className="w-48">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="json">JSON (Recommended)</SelectItem>
                    <SelectItem value="csv">CSV (Limited data)</SelectItem>
                    <SelectItem value="excel">Excel (Limited data)</SelectItem>
                  </SelectContent>
                </Select>
                <p className="text-xs text-muted-foreground">
                  JSON format preserves all data relationships and is recommended for imports
                </p>
              </div>

              <Separator />

              {/* Data Selection */}
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <Label className="text-base">Data to Export</Label>
                  <div className="text-sm text-muted-foreground">
                    Total size: ~{totalExportSize.toFixed(1)} MB
                  </div>
                </div>
                
                <div className="space-y-3">
                  {exportItems.map((item) => (
                    <div key={item.id} className="flex items-start space-x-3 p-3 border rounded-lg">
                      <Checkbox
                        id={item.id}
                        checked={item.included}
                        onCheckedChange={() => toggleExportItem(item.id)}
                        disabled={item.required}
                      />
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <item.icon className="h-4 w-4 text-muted-foreground" />
                          <Label htmlFor={item.id} className="font-medium">
                            {item.name}
                          </Label>
                          {item.required && (
                            <Badge variant="secondary" className="text-xs">Required</Badge>
                          )}
                          <div className="ml-auto text-xs text-muted-foreground">
                            {item.size}
                          </div>
                        </div>
                        <p className="text-sm text-muted-foreground mt-1">
                          {item.description}
                        </p>
                      </div>
                    </div>
                  ))}
                </div>
              </div>

              {/* Export Button */}
              <div className="flex items-center justify-between pt-4 border-t">
                <div className="text-sm text-muted-foreground">
                  {exportItems.filter(item => item.included).length} of {exportItems.length} data types selected
                </div>
                <Button 
                  onClick={handleExport} 
                  disabled={isExporting || exportItems.filter(item => item.included).length === 0}
                >
                  {isExporting ? (
                    <>
                      <RefreshCw className="mr-2 h-4 w-4 animate-spin" />
                      Exporting...
                    </>
                  ) : (
                    <>
                      <Download className="mr-2 h-4 w-4" />
                      Export Data
                    </>
                  )}
                </Button>
              </div>

              {/* Export Progress */}
              {isExporting && (
                <div className="space-y-2">
                  <div className="flex justify-between text-sm">
                    <span>Export Progress</span>
                    <span>{Math.round(exportProgress)}%</span>
                  </div>
                  <Progress value={exportProgress} className="h-2" />
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="import" className="space-y-6">
          {/* File Selection */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Upload className="h-5 w-5" />
                Import Data
              </CardTitle>
              <CardDescription>
                Select an export file to restore your {level} data
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="import-file">Select Export File</Label>
                  <Input
                    id="import-file"
                    type="file"
                    accept=".json,.csv,.xlsx"
                    onChange={handleFileSelect}
                    disabled={isImporting}
                  />
                  <p className="text-xs text-muted-foreground">
                    Supported formats: JSON, CSV, Excel (.json, .csv, .xlsx)
                  </p>
                </div>

                {selectedFile && (
                  <div className="flex items-center gap-3 p-3 bg-muted rounded-lg">
                    <File className="h-8 w-8 text-muted-foreground" />
                    <div className="flex-1">
                      <div className="font-medium">{selectedFile.name}</div>
                      <div className="text-sm text-muted-foreground">
                        {(selectedFile.size / 1024 / 1024).toFixed(2)} MB
                      </div>
                    </div>
                    <Button variant="outline" size="sm" onClick={previewImport}>
                      Preview
                    </Button>
                  </div>
                )}
              </div>

              <Alert>
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>
                  <strong>Important:</strong> Importing data will overwrite existing configurations and settings. 
                  Consider exporting your current data as a backup before proceeding.
                </AlertDescription>
              </Alert>

              <div className="flex items-center gap-2 pt-4 border-t">
                <Button 
                  onClick={handleImport} 
                  disabled={!selectedFile || isImporting}
                >
                  {isImporting ? (
                    <>
                      <RefreshCw className="mr-2 h-4 w-4 animate-spin" />
                      Importing...
                    </>
                  ) : (
                    <>
                      <Upload className="mr-2 h-4 w-4" />
                      Start Import
                    </>
                  )}
                </Button>
                
                {selectedFile && !isImporting && (
                  <Button variant="outline" onClick={previewImport}>
                    <FileText className="mr-2 h-4 w-4" />
                    Preview Import
                  </Button>
                )}
              </div>

              {/* Import Progress */}
              {isImporting && (
                <div className="space-y-2">
                  <div className="flex justify-between text-sm">
                    <span>Import Progress</span>
                    <span>{Math.round(importProgress)}%</span>
                  </div>
                  <Progress value={importProgress} className="h-2" />
                </div>
              )}

              {/* Import Results */}
              {importResult && (
                <Card className={cn(
                  "border",
                  importResult.success ? "border-green-200" : "border-red-200"
                )}>
                  <CardHeader className="pb-3">
                    <CardTitle className="flex items-center gap-2 text-lg">
                      {importResult.success ? (
                        <CheckCircle className="h-5 w-5 text-green-500" />
                      ) : (
                        <AlertCircle className="h-5 w-5 text-red-500" />
                      )}
                      Import {importResult.success ? 'Completed' : 'Failed'}
                    </CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div className="grid grid-cols-2 gap-4 text-sm">
                      <div>
                        <div className="font-medium text-green-600">
                          {importResult.imported} Imported
                        </div>
                        <div className="text-muted-foreground">Items successfully imported</div>
                      </div>
                      {importResult.skipped > 0 && (
                        <div>
                          <div className="font-medium text-yellow-600">
                            {importResult.skipped} Skipped
                          </div>
                          <div className="text-muted-foreground">Items already exist</div>
                        </div>
                      )}
                    </div>

                    {importResult.errors.length > 0 && (
                      <div className="space-y-2">
                        <div className="font-medium text-red-600 text-sm">Errors:</div>
                        <div className="space-y-1">
                          {importResult.errors.map((error, index) => (
                            <div key={index} className="text-sm text-red-600 bg-red-50 p-2 rounded">
                              {error}
                            </div>
                          ))}
                        </div>
                      </div>
                    )}
                  </CardContent>
                </Card>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* Import Preview Dialog */}
      <Dialog open={isPreviewOpen} onOpenChange={setIsPreviewOpen}>
        <DialogContent className="sm:max-w-[600px]">
          <DialogHeader>
            <DialogTitle>Import Preview</DialogTitle>
            <DialogDescription>
              Preview of data that will be imported from your selected file
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div className="p-4 bg-muted rounded-lg">
              <div className="font-medium">File Information</div>
              <div className="text-sm text-muted-foreground mt-1">
                <div>Name: {selectedFile?.name}</div>
                <div>Size: {selectedFile ? (selectedFile.size / 1024 / 1024).toFixed(2) : 0} MB</div>
                <div>Format: {selectedFile?.name.split('.').pop()?.toUpperCase()}</div>
              </div>
            </div>

            <div className="space-y-3">
              <div className="font-medium">Data Types Found:</div>
              {EXPORT_ITEMS.slice(0, 4).map((item) => (
                <div key={item.id} className="flex items-center gap-3 p-2 border rounded">
                  <item.icon className="h-4 w-4 text-muted-foreground" />
                  <div className="flex-1">
                    <div className="text-sm font-medium">{item.name}</div>
                    <div className="text-xs text-muted-foreground">{item.description}</div>
                  </div>
                  <Badge variant="outline" className="text-xs">
                    {Math.floor(Math.random() * 100) + 10} items
                  </Badge>
                </div>
              ))}
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setIsPreviewOpen(false)}>
              Close
            </Button>
            <Button onClick={() => { setIsPreviewOpen(false); handleImport(); }}>
              Proceed with Import
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}