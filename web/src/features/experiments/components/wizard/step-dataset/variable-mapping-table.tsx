'use client'

import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { useExperimentWizard } from '../../../context/experiment-wizard-context'
import type { VariableMappingSource, ExperimentVariableMapping } from '../../../types'

export function VariableMappingTable() {
  const { state, updateDatasetState } = useExperimentWizard()
  const { configState, datasetState } = state
  const { promptVariables } = configState
  const { variableMapping, datasetFields } = datasetState

  if (!datasetFields) {
    return null
  }

  const updateMapping = (variableName: string, source: VariableMappingSource, fieldPath: string) => {
    const existingMapping = variableMapping.find((m) => m.variable_name === variableName)
    const newMapping: ExperimentVariableMapping = {
      variable_name: variableName,
      source,
      field_path: fieldPath,
      is_auto_mapped: false, // User manually changed it
    }

    if (existingMapping) {
      updateDatasetState({
        variableMapping: variableMapping.map((m) =>
          m.variable_name === variableName ? newMapping : m
        ),
      })
    } else {
      updateDatasetState({
        variableMapping: [...variableMapping, newMapping],
      })
    }
  }

  const getFieldValue = (variableName: string): string => {
    const mapping = variableMapping.find((m) => m.variable_name === variableName)
    if (mapping) {
      return `${mapping.source}:${mapping.field_path}`
    }
    return ''
  }

  const handleValueChange = (variableName: string, value: string) => {
    const [source, ...pathParts] = value.split(':')
    const fieldPath = pathParts.join(':')
    updateMapping(variableName, source as VariableMappingSource, fieldPath)
  }

  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-[200px]">Variable</TableHead>
            <TableHead>Map to Dataset Field</TableHead>
            <TableHead className="w-[100px] text-center">Status</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {promptVariables.map((variable) => {
            const mapping = variableMapping.find((m) => m.variable_name === variable)
            const isMapped = mapping && mapping.field_path

            return (
              <TableRow key={variable}>
                <TableCell className="font-mono text-sm">
                  <Badge variant="secondary">{`{{${variable}}}`}</Badge>
                </TableCell>
                <TableCell>
                  <Select value={getFieldValue(variable)} onValueChange={(v) => handleValueChange(variable, v)}>
                    <SelectTrigger className="w-full">
                      <SelectValue placeholder="Select field..." />
                    </SelectTrigger>
                    <SelectContent>
                      {datasetFields.input_fields.length > 0 && (
                        <SelectGroup>
                          <SelectLabel>Input Fields</SelectLabel>
                          {datasetFields.input_fields.map((field) => (
                            <SelectItem
                              key={`input:${field.path}`}
                              value={`dataset_input:${field.path}`}
                            >
                              <div className="flex items-center gap-2">
                                <span>{field.path}</span>
                                <span className="text-xs text-muted-foreground">({field.type})</span>
                              </div>
                            </SelectItem>
                          ))}
                        </SelectGroup>
                      )}
                      {datasetFields.expected_fields.length > 0 && (
                        <SelectGroup>
                          <SelectLabel>Expected Fields</SelectLabel>
                          {datasetFields.expected_fields.map((field) => (
                            <SelectItem
                              key={`expected:${field.path}`}
                              value={`dataset_expected:${field.path}`}
                            >
                              <div className="flex items-center gap-2">
                                <span>{field.path}</span>
                                <span className="text-xs text-muted-foreground">({field.type})</span>
                              </div>
                            </SelectItem>
                          ))}
                        </SelectGroup>
                      )}
                      {datasetFields.metadata_fields.length > 0 && (
                        <SelectGroup>
                          <SelectLabel>Metadata Fields</SelectLabel>
                          {datasetFields.metadata_fields.map((field) => (
                            <SelectItem
                              key={`metadata:${field.path}`}
                              value={`dataset_metadata:${field.path}`}
                            >
                              <div className="flex items-center gap-2">
                                <span>{field.path}</span>
                                <span className="text-xs text-muted-foreground">({field.type})</span>
                              </div>
                            </SelectItem>
                          ))}
                        </SelectGroup>
                      )}
                    </SelectContent>
                  </Select>
                </TableCell>
                <TableCell className="text-center">
                  {isMapped ? (
                    <Badge
                      variant={mapping.is_auto_mapped ? 'secondary' : 'default'}
                      className={mapping.is_auto_mapped ? '' : 'bg-green-500 hover:bg-green-600'}
                    >
                      {mapping.is_auto_mapped ? 'Auto' : 'Mapped'}
                    </Badge>
                  ) : (
                    <Badge variant="outline" className="text-muted-foreground">
                      Unmapped
                    </Badge>
                  )}
                </TableCell>
              </TableRow>
            )
          })}
        </TableBody>
      </Table>
    </div>
  )
}
