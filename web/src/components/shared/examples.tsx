/**
 * Example Usage of Shared Components
 * 
 * This file demonstrates how to use the shared components library.
 * These examples can be used as reference or copied into your components.
 */

import * as React from 'react'
import { ColumnDef } from '@tanstack/react-table'
import { Activity, Users, DollarSign, TrendingUp } from 'lucide-react'

// Import all shared components
import {
  DataTable,
  DataTableColumnHeader,
  MetricCard,
  StatsGrid,
  ProgressRing,
  StatusIndicator,
  EnhancedInput,
  SearchCombobox,
  MultiSelect,
  DateTimePicker,
  SearchInput,
  FilterBar,
  QuickFilters,
  useFilters,
  useQuickFilter,
} from '@/components/shared'

// Example data types
interface User {
  id: string
  name: string
  email: string
  role: string
  status: 'active' | 'inactive' | 'pending'
  lastLogin: Date
}

interface Metric {
  id: string
  title: string
  value: string | number
  icon?: React.ComponentType<{ className?: string }>
  trend?: {
    value: number
    label: string
    direction: 'up' | 'down'
  }
}

// Example 1: Advanced Data Table
export function ExampleUsersTable() {
  const [users] = React.useState<User[]>([
    {
      id: '1',
      name: 'John Doe',
      email: 'john@example.com',
      role: 'admin',
      status: 'active',
      lastLogin: new Date(),
    },
    // ... more users
  ])

  const columns: ColumnDef<User>[] = [
    {
      accessorKey: 'name',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Name" />
      ),
    },
    {
      accessorKey: 'email',
      header: 'Email',
    },
    {
      accessorKey: 'status',
      header: 'Status',
      cell: ({ row }) => {
        const status = row.getValue('status') as string
        return <StatusIndicator status={status as any} />
      },
    },
  ]

  const filterFields = [
    {
      key: 'status',
      title: 'Status',
      options: [
        { label: 'Active', value: 'active' },
        { label: 'Inactive', value: 'inactive' },
        { label: 'Pending', value: 'pending' },
      ],
    },
    {
      key: 'role',
      title: 'Role',
      options: [
        { label: 'Admin', value: 'admin' },
        { label: 'User', value: 'user' },
        { label: 'Viewer', value: 'viewer' },
      ],
    },
  ]

  return (
    <div className="space-y-4">
      <h2 className="text-2xl font-bold">Users Management</h2>
      
      <DataTable
        columns={columns}
        data={users}
        searchField="name"
        filterFields={filterFields}
        onRowAction={(action, user) => {
          console.log('Action:', action, 'User:', user)
          // Handle actions like edit, delete, view
        }}
      />
    </div>
  )
}

// Example 2: Dashboard with Metrics
export function ExampleDashboard() {
  const metrics: Metric[] = [
    {
      id: 'requests',
      title: 'Total Requests',
      value: 12345,
      icon: Activity,
      trend: { value: 12.5, label: 'from last month', direction: 'up' },
    },
    {
      id: 'users',
      title: 'Active Users',
      value: 2350,
      icon: Users,
      trend: { value: 5.2, label: 'from last week', direction: 'up' },
    },
    {
      id: 'revenue',
      title: 'Revenue',
      value: '$45,231',
      icon: DollarSign,
      trend: { value: 2.1, label: 'from last month', direction: 'down' },
    },
    {
      id: 'growth',
      title: 'Growth Rate',
      value: '23.1%',
      icon: TrendingUp,
      trend: { value: 8.3, label: 'from last quarter', direction: 'up' },
    },
  ]

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold">Dashboard Overview</h2>
      
      {/* Metrics Grid */}
      <StatsGrid metrics={metrics} columns={4} />
      
      {/* Progress Indicators */}
      <div className="flex items-center gap-6">
        <div className="text-center space-y-2">
          <ProgressRing progress={75} label="API Health" />
          <p className="text-sm text-muted-foreground">System Status</p>
        </div>
        
        <div className="text-center space-y-2">
          <ProgressRing 
            progress={90} 
            color="success" 
            size="lg" 
            label="Uptime" 
          />
          <p className="text-sm text-muted-foreground">99.9% Uptime</p>
        </div>
        
        <div className="text-center space-y-2">
          <ProgressRing 
            progress={45} 
            color="warning" 
            label="Storage" 
          />
          <p className="text-sm text-muted-foreground">Disk Usage</p>
        </div>
      </div>
    </div>
  )
}

// Example 3: Advanced Form
export function ExampleUserForm() {
  const [formData, setFormData] = React.useState({
    name: '',
    email: '',
    password: '',
    role: '',
    permissions: [],
    startDate: undefined as Date | undefined,
    bio: '',
  })

  const [errors, setErrors] = React.useState<Record<string, string>>({})

  const roleOptions = [
    { value: 'admin', label: 'Administrator', description: 'Full system access' },
    { value: 'manager', label: 'Manager', description: 'Team management access' },
    { value: 'user', label: 'User', description: 'Standard user access' },
    { value: 'viewer', label: 'Viewer', description: 'Read-only access' },
  ]

  const permissionOptions = [
    { value: 'read', label: 'Read' },
    { value: 'write', label: 'Write' },
    { value: 'delete', label: 'Delete' },
    { value: 'admin', label: 'Admin' },
  ]

  return (
    <div className="max-w-2xl space-y-6">
      <h2 className="text-2xl font-bold">Create User</h2>
      
      <form className="space-y-4">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <EnhancedInput
            label="Full Name"
            value={formData.name}
            onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
            error={errors.name}
            required
            placeholder="John Doe"
          />
          
          <EnhancedInput
            label="Email Address"
            type="email"
            value={formData.email}
            onChange={(e) => setFormData(prev => ({ ...prev, email: e.target.value }))}
            error={errors.email}
            required
            placeholder="john@example.com"
          />
        </div>
        
        <EnhancedInput
          label="Password"
          type="password"
          value={formData.password}
          onChange={(e) => setFormData(prev => ({ ...prev, password: e.target.value }))}
          showPasswordToggle
          required
          description="Must be at least 8 characters long"
        />
        
        <SearchCombobox
          label="Role"
          options={roleOptions}
          value={formData.role}
          onValueChange={(value) => setFormData(prev => ({ ...prev, role: value }))}
          placeholder="Select a role..."
          required
        />
        
        <MultiSelect
          label="Permissions"
          options={permissionOptions}
          value={formData.permissions}
          onValueChange={(value) => setFormData(prev => ({ ...prev, permissions: value }))}
          placeholder="Select permissions..."
        />
        
        <DateTimePicker
          label="Start Date"
          value={formData.startDate}
          onValueChange={(date) => setFormData(prev => ({ ...prev, startDate: date }))}
          showTime
        />
        
        <EnhancedTextarea
          label="Bio"
          value={formData.bio}
          onChange={(e) => setFormData(prev => ({ ...prev, bio: e.target.value }))}
          placeholder="Tell us about yourself..."
          showCharCount
          maxLength={500}
          description="Brief description of the user"
        />
      </form>
    </div>
  )
}

// Example 4: Search and Filters
export function ExampleSearchFilters() {
  const [searchValue, setSearchValue] = React.useState('')
  const { activeFilters, addFilter, removeFilter, clearFilters } = useFilters()
  const { activeFilter, setActiveFilter } = useQuickFilter()

  const filterOptions = [
    {
      key: 'status',
      label: 'Status',
      type: 'select' as const,
      options: [
        { value: 'active', label: 'Active' },
        { value: 'inactive', label: 'Inactive' },
        { value: 'pending', label: 'Pending' },
      ],
    },
    {
      key: 'role',
      label: 'Role',
      type: 'multiselect' as const,
      options: [
        { value: 'admin', label: 'Admin' },
        { value: 'user', label: 'User' },
        { value: 'viewer', label: 'Viewer' },
      ],
    },
  ]

  const quickFilters = [
    { key: 'all', label: 'All Users', count: 150 },
    { key: 'active', label: 'Active', count: 120 },
    { key: 'inactive', label: 'Inactive', count: 25 },
    { key: 'pending', label: 'Pending', count: 5 },
  ]

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold">Search & Filter Example</h2>
      
      {/* Search Input */}
      <SearchInput
        value={searchValue}
        onValueChange={setSearchValue}
        onSearch={(value) => console.log('Searching for:', value)}
        placeholder="Search users, emails, or roles..."
        className="max-w-md"
      />
      
      {/* Quick Filters */}
      <div>
        <h3 className="text-sm font-medium mb-2">Quick Filters</h3>
        <QuickFilters
          filters={quickFilters}
          activeFilter={activeFilter}
          onFilterChange={setActiveFilter}
          variant="tabs"
          showCounts
        />
      </div>
      
      {/* Advanced Filters */}
      <div>
        <h3 className="text-sm font-medium mb-2">Advanced Filters</h3>
        <FilterBar
          filters={filterOptions}
          activeFilters={activeFilters}
          onFilterAdd={(filterKey) => {
            // In a real app, you'd show a modal or dropdown to select filter value
            console.log('Add filter:', filterKey)
          }}
          onFilterRemove={removeFilter}
          onFiltersClear={clearFilters}
        />
      </div>
    </div>
  )
}

// Example 5: Status Indicators
export function ExampleStatusIndicators() {
  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold">Status Indicators</h2>
      
      <div className="space-y-4">
        <div>
          <h3 className="text-sm font-medium mb-2">Badge Variant</h3>
          <div className="flex gap-2">
            <StatusIndicator status="online" />
            <StatusIndicator status="offline" />
            <StatusIndicator status="pending" />
            <StatusIndicator status="error" />
            <StatusIndicator status="warning" />
            <StatusIndicator status="success" />
          </div>
        </div>
        
        <div>
          <h3 className="text-sm font-medium mb-2">Dot Variant</h3>
          <div className="space-y-2">
            <StatusIndicator status="online" variant="dot" />
            <StatusIndicator status="offline" variant="dot" />
            <StatusIndicator status="pending" variant="dot" />
          </div>
        </div>
        
        <div>
          <h3 className="text-sm font-medium mb-2">Pill Variant</h3>
          <div className="flex gap-2">
            <StatusIndicator status="online" variant="pill" size="sm" />
            <StatusIndicator status="error" variant="pill" size="md" />
            <StatusIndicator status="success" variant="pill" size="lg" />
          </div>
        </div>
      </div>
    </div>
  )
}