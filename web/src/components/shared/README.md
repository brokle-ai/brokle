# Shared Components Library

A comprehensive collection of reusable UI components built on top of shadcn/ui, designed following the patterns from shadcn-admin for the Brokle dashboard.

## 📦 Components Overview

### 🗂️ Data Tables
- **DataTable** - Advanced table component with sorting, filtering, pagination
- **DataTableToolbar** - Search and filter controls for tables
- **DataTablePagination** - Pagination controls with customizable page sizes
- **DataTableColumnHeader** - Sortable column headers with dropdown actions
- **DataTableFacetedFilter** - Multi-select filter for table columns
- **DataTableViewOptions** - Column visibility controls
- **DataTableRowActions** - Customizable row action menus

### 📊 Metrics & Analytics
- **MetricCard** - Display key metrics with icons, trends, and states
- **StatsGrid** - Grid layout for multiple metric cards
- **ProgressRing** - Circular progress indicators with customizable styling
- **StatusIndicator** - Status badges and indicators with various styles

### 📝 Enhanced Forms
- **FormField** - Enhanced form field wrapper with labels, descriptions, errors
- **EnhancedInput** - Input with password toggle, validation states
- **EnhancedTextarea** - Textarea with character count and validation
- **SearchCombobox** - Searchable select dropdown
- **MultiSelect** - Multi-selection component with badges
- **DateTimePicker** - Date and time selection with calendar

### 🔍 Search & Filters
- **SearchInput** - Debounced search input with clear functionality
- **FilterBar** - Manage multiple active filters with badges
- **QuickFilters** - Preset filter buttons for common use cases

## 🚀 Quick Start

```tsx
import { DataTable, MetricCard, SearchInput } from '@/components/shared'

// Basic table usage
<DataTable
  columns={columns}
  data={data}
  searchField="name"
  filterFields={filterFields}
/>

// Metric card with trend
<MetricCard
  title="Total Requests"
  value={12345}
  icon={Activity}
  trend={{ value: 12.5, label: "from last month", direction: "up" }}
/>

// Search with debouncing
<SearchInput
  value={searchValue}
  onValueChange={setSearchValue}
  onSearch={handleSearch}
  placeholder="Search users..."
/>
```

## 📖 Detailed Usage

### DataTable

The DataTable component provides a powerful, flexible table solution:

```tsx
import { DataTable, DataTableColumnHeader } from '@/components/shared/tables'
import { ColumnDef } from '@tanstack/react-table'

interface User {
  id: string
  name: string
  email: string
  status: 'active' | 'inactive'
}

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
    ],
  },
]

function UsersTable() {
  return (
    <DataTable
      columns={columns}
      data={users}
      searchField="name"
      filterFields={filterFields}
      onRowAction={(action, user) => {
        console.log('Action:', action, 'User:', user)
      }}
    />
  )
}
```

### Metrics Components

Create dashboards with metric cards and grids:

```tsx
import { MetricCard, StatsGrid, ProgressRing } from '@/components/shared/metrics'
import { DollarSign, Users, Activity, TrendingUp } from 'lucide-react'

const metrics = [
  {
    id: 'revenue',
    title: 'Total Revenue',
    value: '$45,231',
    icon: DollarSign,
    trend: { value: 20.1, label: 'from last month', direction: 'up' as const },
  },
  {
    id: 'users',
    title: 'Active Users',
    value: 2350,
    icon: Users,
    trend: { value: 180.1, label: 'from last month', direction: 'up' as const },
  },
]

function Dashboard() {
  return (
    <div className="space-y-6">
      <StatsGrid metrics={metrics} columns={4} />
      
      <div className="flex items-center gap-4">
        <ProgressRing progress={75} label="Complete" />
        <ProgressRing 
          progress={90} 
          color="success" 
          size="lg" 
          label="Health Score" 
        />
      </div>
    </div>
  )
}
```

### Enhanced Forms

Build complex forms with validation and enhanced UX:

```tsx
import { 
  EnhancedInput, 
  SearchCombobox, 
  MultiSelect, 
  DateTimePicker 
} from '@/components/shared/forms'

const userRoles = [
  { value: 'admin', label: 'Administrator' },
  { value: 'user', label: 'User' },
  { value: 'viewer', label: 'Viewer' },
]

function UserForm() {
  return (
    <form className="space-y-4">
      <EnhancedInput
        label="Email"
        type="email"
        required
        error={errors.email}
        placeholder="user@example.com"
      />
      
      <EnhancedInput
        label="Password"
        type="password"
        showPasswordToggle
        required
      />
      
      <SearchCombobox
        label="Role"
        options={userRoles}
        value={selectedRole}
        onValueChange={setSelectedRole}
      />
      
      <MultiSelect
        label="Permissions"
        options={permissions}
        value={selectedPermissions}
        onValueChange={setSelectedPermissions}
      />
      
      <DateTimePicker
        label="Start Date"
        showTime
        value={startDate}
        onValueChange={setStartDate}
      />
    </form>
  )
}
```

### Search & Filters

Implement search and filtering functionality:

```tsx
import { 
  SearchInput, 
  FilterBar, 
  QuickFilters, 
  useFilters, 
  useQuickFilter 
} from '@/components/shared/search'

const filterOptions = [
  {
    key: 'status',
    label: 'Status',
    type: 'select' as const,
    options: [
      { value: 'active', label: 'Active' },
      { value: 'inactive', label: 'Inactive' },
    ],
  },
]

const quickFilters = [
  { key: 'all', label: 'All Users', count: 150 },
  { key: 'active', label: 'Active', count: 120 },
  { key: 'inactive', label: 'Inactive', count: 30 },
]

function UserSearch() {
  const [searchValue, setSearchValue] = useState('')
  const { activeFilters, addFilter, removeFilter, clearFilters } = useFilters()
  const { activeFilter, setActiveFilter } = useQuickFilter()

  return (
    <div className="space-y-4">
      <SearchInput
        value={searchValue}
        onValueChange={setSearchValue}
        placeholder="Search users..."
      />
      
      <QuickFilters
        filters={quickFilters}
        activeFilter={activeFilter}
        onFilterChange={setActiveFilter}
      />
      
      <FilterBar
        filters={filterOptions}
        activeFilters={activeFilters}
        onFilterAdd={addFilter}
        onFilterRemove={removeFilter}
        onFiltersClear={clearFilters}
      />
    </div>
  )
}
```

## 🎨 Styling & Customization

All components are built with Tailwind CSS and support:

- **Custom className props** for styling overrides
- **shadcn/ui design tokens** for consistent theming
- **Responsive design** with mobile-first approach
- **Dark mode support** via CSS variables
- **Accessibility features** with proper ARIA labels

## 📱 Responsive Design

Components are designed mobile-first:

```tsx
// DataTable automatically adapts with responsive columns
<DataTable 
  columns={columns} 
  data={data}
  // Toolbar stacks on mobile, horizontal on desktop
/>

// StatsGrid adjusts columns based on screen size
<StatsGrid 
  metrics={metrics}
  columns={4} // 1 on mobile, 2 on tablet, 4 on desktop
/>

// QuickFilters wrap on smaller screens
<QuickFilters 
  filters={filters}
  variant="tabs" // Converts to dropdown on mobile
/>
```

## 🔧 TypeScript Support

All components are fully typed with TypeScript:

```tsx
interface User {
  id: string
  name: string
  email: string
}

// Full type safety for table columns
const columns: ColumnDef<User>[] = [...]

// Type-safe metric data
interface Metric {
  id: string
  title: string
  value: string | number
  // ... other properties
}
```

## ⚡ Performance

- **Lazy loading** for large datasets in DataTable
- **Debounced search** to reduce API calls
- **Memoized components** to prevent unnecessary re-renders
- **Virtual scrolling** for large lists (coming soon)

## 🧪 Testing

Components are designed for testability:

```tsx
import { render, screen } from '@testing-library/react'
import { DataTable } from '@/components/shared'

test('renders table data', () => {
  render(<DataTable columns={columns} data={testData} />)
  expect(screen.getByText('Test User')).toBeInTheDocument()
})
```

## 🤝 Contributing

When adding new components:

1. Follow the existing patterns and structure
2. Include TypeScript interfaces
3. Add proper documentation
4. Ensure responsive design
5. Test accessibility features
6. Follow the naming conventions

## 📚 Dependencies

- **@tanstack/react-table** - Table functionality
- **@radix-ui/react-*** - Primitive components
- **lucide-react** - Icons
- **date-fns** - Date manipulation
- **tailwindcss** - Styling
- **class-variance-authority** - Component variants