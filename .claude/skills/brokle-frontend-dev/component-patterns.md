# Component Patterns

UI component patterns and templates for Brokle frontend development.

## Existing Shared Components

**Check existing components before creating new ones:**

- `components/ui/*` - shadcn/ui primitives
- `components/shared/*` - Domain-agnostic reusable components
- `components/shared/metrics/` - MetricCard, StatsGrid
- `components/layout/*` - Header, Sidebar, Footer

## Metric Display

Use pre-built `MetricCard` and `StatsGrid`:

```tsx
import { MetricCard, StatsGrid } from '@/components/shared/metrics'

// Single metric
<MetricCard
  title="Total Requests"
  value="1.2M"
  icon={Activity}
  trend={{ value: 12, label: 'vs yesterday', direction: 'up' }}
  loading={isLoading}
/>

// Grid of metrics
<StatsGrid columns={4} gap="md">
  <MetricCard title="Requests" value="1.2M" />
  <MetricCard title="Latency" value="245ms" />
</StatsGrid>
```

Features: Loading skeletons, error handling, trends, icons, auto-formatting.

## Dashboard Layout

```tsx
export function DashboardPage() {
  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Dashboard</h1>
        <Button>Create Project</Button>
      </div>

      {/* Metrics */}
      <StatsGrid columns={4} gap="md">
        <MetricCard title="Requests" value="1.2M" icon={Activity} />
        <MetricCard title="Latency" value="245ms" icon={Clock} />
      </StatsGrid>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card className="p-6"><h3>Request Volume</h3></Card>
        <Card className="p-6"><h3>Latency Distribution</h3></Card>
      </div>
    </div>
  )
}
```

## Data Table

```tsx
<Card>
  <Table>
    <TableHeader>
      <TableRow className="bg-neutral-50 border-b">
        <TableHead className="font-semibold">Column</TableHead>
      </TableRow>
    </TableHeader>
    <TableBody>
      {data.map((row) => (
        <TableRow className="hover:bg-neutral-50 transition-colors border-b last:border-0">
          {/* Cells */}
        </TableRow>
      ))}
    </TableBody>
  </Table>
</Card>
```

## Modal/Dialog

```tsx
<Dialog open={open} onOpenChange={onOpenChange}>
  <DialogContent className="sm:max-w-md">
    <DialogHeader>
      <DialogTitle>Are you sure?</DialogTitle>
      <DialogDescription>This action cannot be undone.</DialogDescription>
    </DialogHeader>
    <DialogFooter className="gap-2 sm:gap-0">
      <Button variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button>
      <Button variant="destructive" onClick={onConfirm}>Delete</Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
```

## State Patterns

### Loading State

```tsx
if (isLoading) return <TableSkeleton />
if (error) return <ErrorState error={error} />
if (!data?.length) return <EmptyState />
return <Table data={data} />
```

### Error State

```tsx
<Card className="p-8 text-center">
  <AlertCircle className="h-12 w-12 text-destructive mx-auto mb-4" />
  <h3 className="text-lg font-semibold mb-2">Something went wrong</h3>
  <p className="text-sm text-muted-foreground mb-4">{error.message}</p>
  <Button onClick={onRetry}>Try Again</Button>
</Card>
```

### Empty State

```tsx
<Card className="p-12 text-center">
  <Inbox className="h-16 w-16 text-muted-foreground mx-auto mb-4" />
  <h3 className="text-xl font-semibold mb-2">{title}</h3>
  <p className="text-muted-foreground mb-6">{description}</p>
  {action}
</Card>
```

## Form Patterns

### Inline Validation

```tsx
<Form {...form}>
  <FormField
    name="email"
    render={({ field }) => (
      <FormItem>
        <FormLabel>Email</FormLabel>
        <FormControl>
          <Input {...field} type="email" />
        </FormControl>
        <FormMessage />
      </FormItem>
    )}
  />
</Form>
```

### Loading Button

```tsx
<Button disabled={isLoading}>
  {isLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
  {isLoading ? 'Saving...' : 'Save'}
</Button>
```

## Toast Notifications

```tsx
import { toast } from 'sonner'

toast.success('Changes saved')

toast.error('Failed to delete', {
  description: 'Permission denied',
  action: { label: 'Retry', onClick: retryDelete },
})

toast.promise(saveProject(), {
  loading: 'Saving...',
  success: 'Saved!',
  error: 'Failed to save',
})
```

## Accessibility (WCAG AA)

1. **Contrast**: 4.5:1 for normal text, 3:1 for large text
2. **Keyboard**: All interactive elements accessible via keyboard
3. **Screen Readers**: Semantic HTML + ARIA labels
4. **Focus**: Visible focus indicators

```tsx
<button aria-label="Close dialog" onClick={onClose}>
  <X className="h-4 w-4" />
</button>

<Button className="focus:ring-2 focus:ring-ring focus:ring-offset-2">
  Visible Focus
</Button>
```

## Chart Pattern

```tsx
import { AreaChart, Area, CartesianGrid, XAxis, YAxis } from 'recharts'

const chartColors = [
  'hsl(var(--chart-1))',
  'hsl(var(--chart-2))',
]

<Card className="p-6">
  <h3 className="text-sm font-medium text-muted-foreground mb-4">Requests</h3>
  <AreaChart width={600} height={200} data={data}>
    <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" />
    <XAxis dataKey="date" stroke="hsl(var(--muted-foreground))" />
    <Area type="monotone" dataKey="value" stroke={chartColors[0]} fill={chartColors[0]} fillOpacity={0.2} />
  </AreaChart>
</Card>
```
