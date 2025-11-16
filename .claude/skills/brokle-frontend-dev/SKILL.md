---
name: brokle-frontend-dev
description: Use this skill when developing, implementing, or modifying Next.js/React frontend code for the Brokle web application. This includes creating components, pages, hooks, API clients, stores, or any frontend features. Invoke this skill at the start of frontend development tasks.
---

# Brokle Frontend Development Skill

This skill provides comprehensive guidance for Next.js/React frontend development following Brokle's feature-based architecture.

## Tech Stack

- **Next.js**: 15.5.2 (App Router, Turbopack)
- **React**: 19.2.0
- **TypeScript**: 5.9.3 (strict mode)
- **Styling**: Tailwind CSS 4.1.15
- **Components**: shadcn/ui
- **State Management**: Zustand (client state) + React Query (server state)
- **Forms**: React Hook Form + Zod validation
- **Testing**: Vitest + React Testing Library + MSW
- **Package Manager**: pnpm
- **Port**: :3000

## Brokle Design System

### Brand Identity
Brokle is an **AI observability and control plane** platform. The design should reflect:
- **Professional & Data-Focused**: Clean, technical aesthetic for developers and engineering teams
- **Trustworthy & Reliable**: Enterprise-grade quality with attention to detail
- **Modern & Innovative**: Forward-thinking AI/ML platform without being flashy
- **Performance-Oriented**: Fast, responsive, purposeful interactions

### Color Palette

**⚠️ AVOID Generic Defaults**: Do NOT use generic purple gradients, default Tailwind colors without customization, or overly saturated accent colors that lack brand identity.

**Color System**: Brokle uses **OKLCH color space** with CSS variables for perceptual uniformity and superior dark mode support.

**Primary Colors** (from `globals.css`):
```css
/* Light mode */
--background: oklch(1 0 0)              /* White */
--foreground: oklch(0.129 0.042 264.695) /* Dark navy text */
--primary: oklch(0.208 0.042 265.755)    /* Primary brand color */
--accent: oklch(0.968 0.007 247.896)     /* Accent color */

/* Dark mode */
--background: oklch(0.129 0.042 264.695) /* Dark navy */
--foreground: oklch(0.984 0.003 247.858) /* Light text */
--primary: oklch(0.929 0.013 255.508)    /* Lighter primary */
--accent: oklch(0.279 0.041 260.031)     /* Darker accent */
```

**Always use CSS variables in components**:
```tsx
// ✅ Correct - uses CSS variables
<div className="bg-primary text-primary-foreground">Primary CTA</div>
<div className="bg-accent text-accent-foreground">Accent section</div>
<div className="text-destructive">Error message</div>

// ❌ Wrong - hardcoded colors
<div className="bg-blue-600 text-white">Button</div>
```

**Semantic Colors** (CSS variables):
- `bg-primary` / `text-primary` - Primary brand actions
- `bg-secondary` / `text-secondary` - Secondary actions
- `bg-accent` / `text-accent` - Accent elements
- `bg-destructive` / `text-destructive` - Error states, delete actions
- `bg-muted` / `text-muted-foreground` - Subtle backgrounds, secondary text
- `border-border` - All borders
- `ring-ring` - Focus rings

**Chart Colors** (CSS variables):
```tsx
// Use chart-1 through chart-5 for data visualization
const chartConfig = {
  colors: [
    'hsl(var(--chart-1))',
    'hsl(var(--chart-2))',
    'hsl(var(--chart-3))',
    'hsl(var(--chart-4))',
    'hsl(var(--chart-5))',
  ],
}
```

**Why OKLCH?**
- Perceptual uniformity (equal brightness changes look equal)
- Better color interpolation (no muddy gradients)
- Easier dark mode (adjust lightness without hue shifts)
- Future-proof (native browser support growing)

### Dark Mode Support

**Implementation**: Brokle uses `next-themes` for dark mode with CSS variable switching.

**Usage in Components**:
```tsx
'use client'
import { useTheme } from 'next-themes'

export function ThemeToggle() {
  const { theme, setTheme } = useTheme()

  return (
    <Button onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}>
      Toggle Theme
    </Button>
  )
}
```

**Dark Mode Classes** (automatic with CSS variables):
```tsx
// Colors automatically switch based on theme
<div className="bg-background text-foreground">
  Adapts to light/dark mode automatically
</div>

// Force dark mode styles (rarely needed)
<div className="dark:bg-gray-900 dark:text-white">
  Custom dark mode override
</div>
```

**Design Principles for Dark Mode**:
1. **Use CSS variables** - Colors switch automatically (`bg-background`, `text-foreground`)
2. **Reduce contrast** - Dark mode uses slightly muted colors (check `globals.css` `.dark` section)
3. **Adjust opacity** - Use `opacity-*` or color opacity (`bg-primary/80`) for subtle effects
4. **Test both modes** - Always verify components work in light and dark themes
5. **Accessible contrast** - Ensure 4.5:1 contrast ratio in both modes

### Background Patterns

**⚠️ AVOID Solid Colors**: Do NOT use plain white or single-color backgrounds for major sections.

**Recommended Patterns**:
1. **Subtle Gradients**: Diagonal gradients with 2-3 stops (e.g., `bg-gradient-to-br from-neutral-50 via-blue-50/30 to-neutral-50`)
2. **Grid Patterns**: Subtle dot or line grids for technical feel (`bg-[url('/grid.svg')]`)
3. **Depth Layers**: Overlapping sections with different opacities
4. **Atmospheric Effects**: Radial gradients for hero sections

### Spacing & Layout
- **Container Max Width**: `1280px` (xl breakpoint)
- **Content Padding**: `px-4 md:px-6 lg:px-8`
- **Section Spacing**: `py-12 md:py-16 lg:py-24`
- **Card Padding**: `p-6 md:p-8`
- **Grid Gaps**: `gap-4 md:gap-6 lg:gap-8`

## Typography Guidelines

### Font Families

**Current Implementation**: Brokle currently uses **Inter** for all text (loaded in `app/layout.tsx`).

**⚠️ AVOID Boring Typography**: While Inter is used currently, avoid defaulting to it without consideration. Use font weights strategically to create visual hierarchy.

**Current Font Stack**:
```typescript
// app/layout.tsx - Currently loaded
import { Inter } from "next/font/google"
const inter = Inter({
  subsets: ["latin"],
  weight: ["100", "200", "300", "400", "500", "600", "700", "800", "900"],
})
```

**Usage in Components**:
- `font-sans` → Inter (all UI text, body copy)
- No separate display or monospace fonts currently loaded

**Aspirational Fonts** (Future Enhancement):
When enhancing the design system, consider these distinctive alternatives:

1. **Display/Headings**: **"Bricolage Grotesque"** or **"Cabinet Grotesk"**
   - Use for: H1, H2, hero sections, marketing headers
   - Characteristics: Distinctive, modern, slightly geometric
   - Weights: 600 (Semibold), 700 (Bold), 800 (Extrabold)
   - **Impact**: Creates unique brand identity vs generic Inter headings

2. **Alternative Body**: **"Manrope"**
   - Use for: Paragraphs, UI text (alternative to Inter)
   - Characteristics: Slightly warmer than Inter, excellent readability
   - Weights: 400 (Regular), 500 (Medium), 600 (Semibold)
   - **Note**: Configured in `tailwind.config.ts` but not loaded

3. **Monospace/Code**: **"JetBrains Mono"** or **"Fira Code"**
   - Use for: Code snippets, API keys, log data, technical values
   - Characteristics: Excellent readability, ligatures
   - Weights: 400 (Regular), 500 (Medium), 600 (Semibold)
   - **Impact**: Better code readability vs system monospace

4. **Data/Metrics**: **"IBM Plex Mono"** (tabular figures)
   - Use for: Numbers in tables, metrics, dashboards
   - Characteristics: Tabular figures, consistent width
   - Weights: 500 (Medium), 600 (Semibold)
   - **Impact**: Aligned numbers in data-heavy interfaces

### Typography Scale

```typescript
// Tailwind config - use these classes
{
  'text-xs': '0.75rem',      // 12px - Labels, captions
  'text-sm': '0.875rem',     // 14px - Secondary text
  'text-base': '1rem',       // 16px - Body text
  'text-lg': '1.125rem',     // 18px - Emphasis
  'text-xl': '1.25rem',      // 20px - Subheadings
  'text-2xl': '1.5rem',      // 24px - H3
  'text-3xl': '1.875rem',    // 30px - H2
  'text-4xl': '2.25rem',     // 36px - H1
  'text-5xl': '3rem',        // 48px - Display/Hero
}
```

### Font Weight Hierarchy

**⚠️ Use High Contrast**: Vary weights dramatically (400 → 700, not 400 → 500) for visual hierarchy.

- **Display**: `font-bold (700)` or `font-extrabold (800)`
- **Headings**: `font-semibold (600)` or `font-bold (700)`
- **Body**: `font-normal (400)` or `font-medium (500)`
- **Emphasis**: `font-semibold (600)`
- **Subtle**: `font-light (300)` - use sparingly

### Typography in Practice

**Current Implementation** (using Inter with high-contrast weights):

```tsx
// Hero section - high contrast with Inter
<h1 className="font-sans text-5xl font-extrabold text-foreground">
  AI Observability Platform
</h1>
<p className="font-sans text-lg font-normal text-muted-foreground">
  Monitor, optimize, and control your AI applications
</p>

// Dashboard metric - use font-semibold or font-bold for numbers
<div className="font-sans text-3xl font-bold text-foreground">
  $1,234.56
</div>

// Code/API key - use font-mono (system monospace currently)
<code className="font-mono text-sm font-medium text-foreground bg-muted px-2 py-1 rounded">
  bk_AbCdEfGhIjKl...
</code>

// Card title - medium weight for subtle hierarchy
<h3 className="font-sans text-sm font-medium text-muted-foreground">
  Total Requests
</h3>
```

**With Aspirational Fonts** (future):
```tsx
// Display font for hero headings (if loaded)
<h1 className="font-display text-5xl font-extrabold text-foreground">
  AI Observability Platform
</h1>

// Tabular font for aligned metrics (if loaded)
<div className="font-tabular text-3xl font-semibold text-foreground">
  $1,234.56
</div>
```

## Visual Design Principles

### Depth & Layering

**⚠️ Create Atmospheric Depth**: Do NOT use flat, single-layer designs. Build depth through layering, shadows, and overlapping elements.

**Techniques**:
1. **Shadows**: Use subtle shadows for elevation
   - `shadow-sm`: Subtle lift (cards on white background)
   - `shadow-md`: Medium elevation (dropdowns, popovers)
   - `shadow-lg`: High elevation (modals, dialogs)
   - `shadow-xl`: Maximum elevation (tooltips, notifications)

2. **Borders**: Use borders sparingly, prefer shadows
   - `border-neutral-200`: Light borders for structure
   - `border-neutral-300`: Medium borders for emphasis
   - `ring-2 ring-accent-500/20`: Focus states

3. **Overlapping Elements**: Create depth with z-index layers
   ```tsx
   <div className="relative">
     <div className="absolute -top-4 -left-4 w-24 h-24 bg-blue-500/10 rounded-full blur-2xl" />
     <Card className="relative z-10">Content</Card>
   </div>
   ```

### Background Treatments

**Examples for Different Sections**:

```tsx
// Hero section - atmospheric gradient
<section className="relative overflow-hidden bg-gradient-to-br from-neutral-900 via-primary-900 to-neutral-900">
  <div className="absolute inset-0 bg-[url('/grid.svg')] opacity-20" />
  <div className="absolute top-0 right-0 w-96 h-96 bg-accent-500/20 rounded-full blur-3xl" />
  <div className="relative z-10">{/* Content */}</div>
</section>

// Dashboard section - subtle depth
<section className="bg-gradient-to-b from-neutral-50 to-white">
  <div className="absolute inset-0 bg-[url('/dots.svg')] opacity-5" />
  {/* Content */}
</section>

// Card with elevation
<Card className="bg-white shadow-lg border border-neutral-100 hover:shadow-xl transition-shadow">
  {/* Content */}
</Card>
```

### Data Visualization Aesthetics

**Chart Design Principles**:
1. **Color Consistency**: Use brand color palette for all charts
2. **Accessibility**: Ensure 4.5:1 contrast ratios
3. **Annotations**: Add clear labels, tooltips, and legends
4. **Responsiveness**: Charts adapt to container width

**Chart Component Pattern** (using Recharts + CSS variables):
```tsx
import { AreaChart, Area, CartesianGrid, XAxis, YAxis } from 'recharts'
import { Card } from '@/components/ui/card'

// Use CSS variables for chart colors
const chartConfig = {
  colors: [
    'hsl(var(--chart-1))',
    'hsl(var(--chart-2))',
    'hsl(var(--chart-3))',
    'hsl(var(--chart-4))',
    'hsl(var(--chart-5))',
  ],
}

export function RequestsChart({ data }: { data: any[] }) {
  return (
    <Card className="p-6">
      <div className="flex items-baseline justify-between mb-4">
        <h3 className="text-sm font-medium text-muted-foreground">Total Requests</h3>
        <span className="text-2xl font-bold text-foreground">1.2M</span>
      </div>
      <AreaChart width={600} height={200} data={data}>
        <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" />
        <XAxis dataKey="date" stroke="hsl(var(--muted-foreground))" />
        <YAxis stroke="hsl(var(--muted-foreground))" />
        <Area
          type="monotone"
          dataKey="requests"
          stroke={chartConfig.colors[0]}
          fill={chartConfig.colors[0]}
          fillOpacity={0.2}
        />
      </AreaChart>
    </Card>
  )
}
```

### Responsive Design Patterns

**Breakpoints** (from Tailwind):
- `sm`: 640px - Small tablets
- `md`: 768px - Tablets
- `lg`: 1024px - Laptops
- `xl`: 1280px - Desktops
- `2xl`: 1536px - Large desktops

**Layout Patterns**:
```tsx
// Stack on mobile, grid on desktop
<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
  <Card>1</Card>
  <Card>2</Card>
  <Card>3</Card>
</div>

// Hide on mobile, show on desktop
<div className="hidden lg:block">Desktop only</div>

// Responsive text sizing
<h1 className="text-3xl md:text-4xl lg:text-5xl font-display font-extrabold">
  Responsive Heading
</h1>
```

## Motion & Animation

**Current Implementation**: Brokle uses **CSS animations** via `tailwindcss-animate` and custom keyframes in `globals.css`.

**⚠️ Prioritize Key Interactions**: Do NOT add animations everywhere. Focus on high-impact moments: page loads, data updates, user feedback.

### Current CSS Animations

**Available Utility Classes** (from `globals.css`):

```tsx
// Utility classes for common animations
<div className="fade-in">Fades in on mount</div>
<div className="slide-in-right">Slides from right</div>
<div className="slide-in-left">Slides from left</div>
<div className="zoom-in">Zooms in smoothly</div>

// Hover effects
<div className="btn-hover">Button with scale on hover</div>
<div className="card-hover">Card lifts on hover</div>

// Continuous animations
<div className="pulse-subtle">Subtle pulsing effect</div>
<div className="bounce-gentle">Gentle bouncing</div>
```

**Built-in Tailwind Animations**:
```tsx
// Pulse (loading states)
<div className="animate-pulse">Loading...</div>

// Spin (loaders)
<Loader2 className="animate-spin" />

// Bounce
<div className="animate-bounce">Notification badge</div>
```

### Micro-Interactions

**Button States** (current implementation):
```tsx
// From components/ui/button.tsx - already implements active:scale-95
<Button className="active:scale-95 hover:shadow-md transition-all duration-200">
  Click Me
</Button>

// Add custom hover effects
<Button className="hover:scale-105 transition-transform">
  Enhanced Hover
</Button>
```

**Card Hover Effects**:
```tsx
// Use existing card-hover utility
<Card className="card-hover">
  Lifts on hover with shadow
</Card>

// Or custom
<Card className="transition-all hover:shadow-lg hover:-translate-y-1">
  Custom lift effect
</Card>
```

**Loading States** - Skeleton screens:
```tsx
// Use existing skeleton component
import { Skeleton } from '@/components/ui/skeleton'

export function CardSkeleton() {
  return (
    <Card className="p-6">
      <div className="space-y-4">
        <Skeleton className="h-4 w-24" />
        <Skeleton className="h-8 w-32" />
        <Skeleton className="h-32 w-full" />
      </div>
    </Card>
  )
}
```

### Transition Patterns

**Modal/Dialog Entry**:
```tsx
// shadcn Dialog has built-in animations
import { Dialog, DialogContent } from '@/components/ui/dialog'

<Dialog>
  <DialogContent className="animate-in fade-in-0 zoom-in-95">
    {/* Content */}
  </DialogContent>
</Dialog>
```

**Toast Notifications**:
```tsx
// Sonner has built-in animations
import { toast } from 'sonner'

toast.success('Saved successfully!')
// Automatically animates in from position
```

**Page Transitions**:
```tsx
// Use CSS utility classes for page load
export function DashboardPage() {
  return (
    <div className="fade-in">
      <h1>Dashboard</h1>
      {/* Content */}
    </div>
  )
}
```

### Advanced Animations (Future Enhancement)

**Framer Motion** (not currently installed - consider for complex animations):

```tsx
// Install: pnpm add framer-motion
'use client'

import { motion } from 'framer-motion'

// Staggered grid reveals
const container = {
  hidden: { opacity: 0 },
  show: {
    opacity: 1,
    transition: { staggerChildren: 0.1 }
  }
}

const item = {
  hidden: { opacity: 0, y: 20 },
  show: { opacity: 1, y: 0 }
}

export function DashboardGrid() {
  return (
    <motion.div
      variants={container}
      initial="hidden"
      animate="show"
      className="grid grid-cols-3 gap-6"
    >
      <motion.div variants={item}><MetricCard /></motion.div>
      <motion.div variants={item}><MetricCard /></motion.div>
      <motion.div variants={item}><MetricCard /></motion.div>
    </motion.div>
  )
}

// Data value changes
<motion.div
  key={value}
  initial={{ scale: 1.2 }}
  animate={{ scale: 1 }}
  transition={{ duration: 0.3 }}
>
  {value}
</motion.div>

// Reduced motion support
const shouldReduceMotion = useReducedMotion()
```

### Performance Considerations

**CSS vs JS Animations**:
- **CSS** (`transition-*`, utility classes): Use for hover, focus, simple state changes (current approach)
- **Framer Motion**: Use for complex orchestrations, staggered animations, gesture interactions (future)

**Always Support Reduced Motion**:
```css
/* Automatically handled in globals.css via prefers-reduced-motion */
@media (prefers-reduced-motion: reduce) {
  * {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
```

## UI/UX Best Practices

### Accessibility (WCAG AA)

**Mandatory Requirements**:
1. **Color Contrast**: Minimum 4.5:1 for normal text, 3:1 for large text
   ```tsx
   // ✅ Good contrast
   <p className="text-neutral-900 bg-white">Readable text</p>

   // ❌ Poor contrast
   <p className="text-neutral-400 bg-neutral-300">Hard to read</p>
   ```

2. **Keyboard Navigation**: All interactive elements accessible via keyboard
   ```tsx
   // Use semantic HTML and shadcn components
   <Button>Accessible by default</Button>
   <Dialog>Traps focus automatically</Dialog>
   ```

3. **Screen Reader Support**: Use semantic HTML and ARIA labels
   ```tsx
   <button aria-label="Close dialog" onClick={onClose}>
     <X className="h-4 w-4" />
   </button>

   <input
     type="text"
     id="email"
     aria-describedby="email-error"
   />
   <p id="email-error" className="text-error text-sm">
     Invalid email address
   </p>
   ```

4. **Focus Indicators**: Always visible focus states
   ```tsx
   <Button className="focus:ring-2 focus:ring-accent-500 focus:ring-offset-2">
     Visible Focus
   </Button>
   ```

### State Patterns

**Loading States**:
```tsx
export function DataTable() {
  const { data, isLoading, error } = useQuery(...)

  if (isLoading) return <TableSkeleton />
  if (error) return <ErrorState error={error} />
  if (!data?.length) return <EmptyState />

  return <Table data={data} />
}
```

**Error States**:
```tsx
export function ErrorState({ error, onRetry }: ErrorStateProps) {
  return (
    <Card className="p-8 text-center">
      <AlertCircle className="h-12 w-12 text-error mx-auto mb-4" />
      <h3 className="text-lg font-semibold text-neutral-900 mb-2">
        Something went wrong
      </h3>
      <p className="text-sm text-neutral-600 mb-4">
        {error.message}
      </p>
      <Button onClick={onRetry}>Try Again</Button>
    </Card>
  )
}
```

**Empty States**:
```tsx
export function EmptyState({ title, description, action }: EmptyStateProps) {
  return (
    <Card className="p-12 text-center">
      <Inbox className="h-16 w-16 text-neutral-400 mx-auto mb-4" />
      <h3 className="text-xl font-semibold text-neutral-900 mb-2">
        {title}
      </h3>
      <p className="text-neutral-600 mb-6 max-w-md mx-auto">
        {description}
      </p>
      {action}
    </Card>
  )
}
```

**Success States** - Toast notifications:
```tsx
import { toast } from 'sonner'

toast.success('Project created successfully', {
  description: 'Your project is ready to use',
})
```

### Form UX Patterns

**Inline Validation**:
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
        {/* Shows error immediately on blur */}
        <FormMessage />
      </FormItem>
    )}
  />
</Form>
```

**Multi-Step Forms**:
```tsx
export function MultiStepForm() {
  const [step, setStep] = useState(1)

  return (
    <Card className="p-8">
      {/* Progress indicator */}
      <div className="flex items-center justify-between mb-8">
        <Step active={step >= 1} />
        <Step active={step >= 2} />
        <Step active={step >= 3} />
      </div>

      {/* Step content */}
      {step === 1 && <StepOne onNext={() => setStep(2)} />}
      {step === 2 && <StepTwo onNext={() => setStep(3)} onBack={() => setStep(1)} />}
      {step === 3 && <StepThree onBack={() => setStep(2)} />}
    </Card>
  )
}
```

**Loading Buttons**:
```tsx
<Button disabled={isLoading}>
  {isLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
  {isLoading ? 'Saving...' : 'Save'}
</Button>
```

### Toast & Notification Patterns

```tsx
import { toast } from 'sonner'

// Success
toast.success('Changes saved')

// Error with action
toast.error('Failed to delete project', {
  description: 'You do not have permission',
  action: {
    label: 'Retry',
    onClick: () => retryDelete(),
  },
})

// Loading with promise
toast.promise(saveProject(), {
  loading: 'Saving project...',
  success: 'Project saved successfully',
  error: 'Failed to save project',
})
```

## Component Design Patterns

### Existing Shared Components

**⚠️ Use Existing Components First**: Brokle has pre-built shared components in `components/shared/`. Always check for existing implementations before creating new patterns.

**Metric Display** (`components/shared/metrics/`):

```tsx
// MetricCard - Pre-built with loading, error, and trend states
import { MetricCard, StatsGrid } from '@/components/shared/metrics'

// Single metric card
<MetricCard
  title="Total Requests"
  value="1.2M"
  description="Last 24 hours"
  icon={ActivityIcon}
  trend={{ value: 12, label: 'vs yesterday', direction: 'up' }}
  loading={isLoading}
  error={error}
/>

// Grid of metrics with responsive columns
<StatsGrid columns={4} gap="md">
  <MetricCard title="Requests" value="1.2M" />
  <MetricCard title="Latency" value="245ms" />
  <MetricCard title="Errors" value="0.3%" />
  <MetricCard title="Cost" value="$1,234" />
</StatsGrid>

// Or pass metrics array
<StatsGrid
  metrics={[
    { id: '1', title: 'Requests', value: '1.2M', trend: {...} },
    { id: '2', title: 'Latency', value: '245ms', trend: {...} },
  ]}
  columns={4}
  loading={isLoading}
/>
```

**Key Features**:
- ✅ Loading skeleton states
- ✅ Error handling
- ✅ Trend indicators (up/down arrows with color)
- ✅ Icon support
- ✅ Responsive grid (1/2/3/4 columns)
- ✅ Number formatting (automatic commas)

**Other Shared Components**:
- `components/ui/*` - shadcn/ui primitives (Button, Card, Dialog, etc.)
- `components/shared/*` - Domain-agnostic reusable components
- `components/layout/*` - App shell (Header, Sidebar, Footer)

### Dashboard Layouts

**Standard Dashboard Grid** (using existing components):
```tsx
import { MetricCard, StatsGrid } from '@/components/shared/metrics'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Activity, Clock, AlertCircle, DollarSign } from 'lucide-react'

export function DashboardPage() {
  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-foreground">Dashboard</h1>
        <Button>Create Project</Button>
      </div>

      {/* Metrics Row - Use StatsGrid */}
      <StatsGrid columns={4} gap="md">
        <MetricCard
          title="Total Requests"
          value="1.2M"
          icon={Activity}
          trend={{ value: 12, label: 'vs yesterday', direction: 'up' }}
        />
        <MetricCard
          title="Avg Latency"
          value="245ms"
          icon={Clock}
          trend={{ value: 8, label: 'improvement', direction: 'down' }}
        />
        <MetricCard
          title="Error Rate"
          value="0.3%"
          icon={AlertCircle}
          trend={{ value: 2, label: 'reduction', direction: 'down' }}
        />
        <MetricCard
          title="Cost"
          value="$1,234"
          icon={DollarSign}
          trend={{ value: 5, label: 'vs last week', direction: 'up' }}
        />
      </StatsGrid>

      {/* Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card className="p-6">
          <h3 className="text-lg font-semibold mb-4">Request Volume</h3>
          {/* Use Recharts or custom chart component */}
        </Card>
        <Card className="p-6">
          <h3 className="text-lg font-semibold mb-4">Latency Distribution</h3>
          {/* Use Recharts or custom chart component */}
        </Card>
      </div>
    </div>
  )
}
```

### Data Table Aesthetics

```tsx
export function DataTable<T>({ data, columns }: DataTableProps<T>) {
  return (
    <Card>
      <Table>
        <TableHeader>
          <TableRow className="bg-neutral-50 border-b border-neutral-200">
            {columns.map(col => (
              <TableHead className="font-semibold text-neutral-900">
                {col.header}
              </TableHead>
            ))}
          </TableRow>
        </TableHeader>
        <TableBody>
          {data.map((row, i) => (
            <TableRow
              key={i}
              className="hover:bg-neutral-50 transition-colors border-b border-neutral-100 last:border-0"
            >
              {/* Cells */}
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </Card>
  )
}
```

### Modal & Dialog Treatments

```tsx
export function ConfirmDialog({ open, onOpenChange, onConfirm }: ConfirmDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="text-xl font-semibold">
            Are you sure?
          </DialogTitle>
          <DialogDescription className="text-neutral-600">
            This action cannot be undone. This will permanently delete your project.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter className="gap-2 sm:gap-0">
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button variant="destructive" onClick={onConfirm}>
            Delete Project
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
```

### Metric Card Implementation

**Use Existing Component**: `components/shared/metrics/metric-card.tsx`

The pre-built `MetricCard` component provides:
- **Props**: `title`, `value`, `description`, `icon`, `trend`, `loading`, `error`
- **Auto-formatting**: Numbers get commas (1234 → 1,234)
- **Loading state**: Automatic skeleton animation
- **Error state**: Red border with error message
- **Trend display**: Up/down arrows with green/red colors
- **Hover effect**: Subtle shadow lift (`.hover:shadow-lg`)

**Source Reference**: See `web/src/components/shared/metrics/metric-card.tsx:28-114` for implementation details.

**Custom Metric Card** (if you need different behavior):
```tsx
export function CustomMetricCard({ title, value }: CustomMetricProps) {
  return (
    <Card className="p-6 hover:shadow-lg transition-shadow">
      <h3 className="text-sm font-medium text-muted-foreground">{title}</h3>
      <p className="text-3xl font-bold text-foreground mt-2">{value}</p>
    </Card>
  )
}
```

## Feature-Based Architecture

### Directory Structure

```
web/src/
├── app/                      # Next.js App Router (routing only)
│   ├── (auth)/              # Auth route group
│   ├── (dashboard)/         # Dashboard routes
│   └── (errors)/            # Error pages
├── features/                # Domain features (self-contained)
│   ├── authentication/      # User auth, sessions, OAuth
│   ├── organizations/       # Org management, members, invitations
│   ├── projects/           # Project dashboard, API keys, settings
│   ├── analytics/          # Usage analytics and metrics
│   ├── billing/            # Billing and subscription management
│   ├── gateway/            # AI gateway configuration
│   ├── settings/           # User settings and preferences
│   └── tasks/              # Task management
├── components/              # Shared components only
│   ├── ui/                 # shadcn/ui primitives
│   ├── layout/             # App shell (header, sidebar, footer)
│   ├── guards/             # Auth guards
│   ├── shared/             # Generic reusable components
│   ├── navigation/         # Navigation components
│   ├── notifications/      # Notification components
│   ├── error-boundary/     # Error boundaries
│   ├── audit/              # Audit components
│   ├── collaboration/      # Collaboration components
│   ├── data/               # Data components
│   ├── templates/          # Template components
│   └── wizard/             # Wizard components
├── lib/                    # Core infrastructure
│   ├── api/core/           # BrokleAPIClient (HTTP client)
│   ├── auth/               # JWT utilities
│   └── utils/              # Pure utilities
├── hooks/                  # Global hooks (use-mobile, etc.)
├── stores/                 # Global stores (ui-store.ts)
├── context/                # Cross-feature context (workspace-context)
├── types/                  # Shared types
├── assets/                 # Static assets (logos, icons, SVGs)
│   ├── brand-icons/        # Provider/brand icons
│   └── custom/             # Custom graphics
├── utils/                  # Small utilities
└── __tests__/              # Test infrastructure (MSW, utilities)
```

### Feature Structure

**Note**: Check `web/src/features/{feature}/index.ts` for current exports and implementation status.

Each feature in `features/[feature]/` has:
- `components/` - Feature-specific UI components
- `hooks/` - Feature-specific React hooks
- `api/` - API functions for this domain
- `stores/` - Zustand stores (optional)
- `types/` - TypeScript definitions
- `__tests__/` - Feature tests
- `index.ts` - Public API exports (ONLY way to import)

## Critical Import Rules (MANDATORY)

### ✅ Allowed Imports

```typescript
// Import from feature public API
import { useAuth, SignInForm } from '@/features/authentication'
import { useOrganization } from '@/features/organizations'

// Shared components
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'

// Utilities
import { cn } from '@/lib/utils'
import { apiClient } from '@/lib/api/core/client'
```

### ❌ Forbidden Imports

```typescript
// DON'T import internal feature files directly
import { useAuth } from '@/features/authentication/hooks/use-auth'  // ❌

// DON'T import from other features' internals
import { AuthStore } from '@/features/authentication/stores/auth-store'  // ❌

// DON'T bypass feature index
import { LoginForm } from '@/features/authentication/components/login-form'  // ❌
```

**Rule**: Always import from `@/features/[feature]` (feature index), never from internal paths.

## Component Patterns

### Server Components (Default)

```typescript
// No 'use client' directive
export default async function Page({ params }: { params: { id: string } }) {
  // Can fetch data server-side
  const data = await fetchData(params.id)

  return (
    <div>
      <ServerComponent data={data} />
      <ClientComponent />
    </div>
  )
}
```

**When to use**:
- Default choice for all components
- Static content
- SEO-friendly pages
- Data fetching at build/request time

### Client Components (When Needed)

```typescript
'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'

export function InteractiveComponent() {
  const [count, setCount] = useState(0)

  return (
    <Button onClick={() => setCount(count + 1)}>
      Count: {count}
    </Button>
  )
}
```

**When to use**:
- Event handlers (onClick, onChange, etc.)
- React hooks (useState, useEffect, useContext)
- Browser APIs (localStorage, window, etc.)
- Third-party libraries requiring browser

## State Management Strategy

### 1. Server State (React Query)
Managed in `context/` providers:

```typescript
// context/workspace-context.tsx
'use client'

import { createContext, useContext } from 'react'
import { useQuery } from '@tanstack/react-query'

export function WorkspaceProvider({ children }: { children: React.Node }) {
  const { data: workspace } = useQuery({
    queryKey: ['workspace'],
    queryFn: fetchWorkspace,
  })

  return (
    <WorkspaceContext.Provider value={{ workspace }}>
      {children}
    </WorkspaceContext.Provider>
  )
}
```

### 2. Client State (Zustand)
Feature-specific stores in `features/[feature]/stores/`:

```typescript
// features/authentication/stores/auth-store.ts
import { create } from 'zustand'

interface AuthState {
  isAuthenticated: boolean
  user: User | null
  setUser: (user: User | null) => void
}

export const useAuthStore = create<AuthState>((set) => ({
  isAuthenticated: false,
  user: null,
  setUser: (user) => set({ user, isAuthenticated: !!user }),
}))
```

### 3. URL State
Use `useSearchParams()` for filters, pagination:

```typescript
'use client'

import { useSearchParams, useRouter } from 'next/navigation'

export function FilterComponent() {
  const searchParams = useSearchParams()
  const router = useRouter()

  const filter = searchParams.get('filter') ?? 'all'

  const setFilter = (newFilter: string) => {
    const params = new URLSearchParams(searchParams)
    params.set('filter', newFilter)
    router.push(`?${params.toString()}`)
  }

  return <div>Filter: {filter}</div>
}
```

### 4. Form State
React Hook Form + Zod validation:

```typescript
'use client'

import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'

const schema = z.object({
  email: z.string().email('Invalid email'),
  password: z.string().min(8, 'At least 8 characters'),
})

type FormData = z.infer<typeof schema>

export function LoginForm() {
  const { register, handleSubmit, formState: { errors } } = useForm<FormData>({
    resolver: zodResolver(schema),
  })

  const onSubmit = (data: FormData) => {
    console.log(data)
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <input {...register('email')} />
      {errors.email && <span>{errors.email.message}</span>}
    </form>
  )
}
```

## Feature Templates

### Feature Index Pattern

```typescript
// features/my-feature/index.ts

// Hooks
export { useMyFeature } from './hooks/use-my-feature'
export { useMyFeatureData } from './hooks/use-my-feature-data'

// Components
export { MyComponent } from './components/my-component'
export { MyForm } from './components/my-form'

// Types (selective export)
export type { MyFeatureData, MyFeatureConfig } from './types'

// DO NOT export internal implementation details
// DO NOT export stores directly (use hooks instead)
```

### Component Template

```typescript
// features/my-feature/components/my-component.tsx
'use client'

import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { useMyFeature } from '../hooks/use-my-feature'

interface MyComponentProps {
  id: string
  onComplete?: () => void
}

export function MyComponent({ id, onComplete }: MyComponentProps) {
  const { data, isLoading, error } = useMyFeature(id)

  if (isLoading) return <div>Loading...</div>
  if (error) return <div>Error: {error.message}</div>

  return (
    <Card>
      <h2 className="text-2xl font-bold">{data?.title}</h2>
      <Button onClick={onComplete}>Complete</Button>
    </Card>
  )
}
```

### API Client Pattern

```typescript
// features/my-feature/api/my-feature-api.ts
import { apiClient } from '@/lib/api/core/client'
import type { MyFeatureResponse, CreateMyFeatureRequest } from '../types'

export async function getMyFeature(id: string): Promise<MyFeatureResponse> {
  const response = await apiClient.get<MyFeatureResponse>(`/my-feature/${id}`)
  return response.data
}

export async function createMyFeature(
  data: CreateMyFeatureRequest
): Promise<MyFeatureResponse> {
  const response = await apiClient.post<MyFeatureResponse>('/my-feature', data)
  return response.data
}

export async function updateMyFeature(
  id: string,
  data: Partial<CreateMyFeatureRequest>
): Promise<MyFeatureResponse> {
  const response = await apiClient.put<MyFeatureResponse>(`/my-feature/${id}`, data)
  return response.data
}

export async function deleteMyFeature(id: string): Promise<void> {
  await apiClient.delete(`/my-feature/${id}`)
}
```

### Hook Pattern

```typescript
// features/my-feature/hooks/use-my-feature.ts
'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getMyFeature, updateMyFeature } from '../api/my-feature-api'

export function useMyFeature(id: string) {
  const queryClient = useQueryClient()

  const query = useQuery({
    queryKey: ['my-feature', id],
    queryFn: () => getMyFeature(id),
  })

  const updateMutation = useMutation({
    mutationFn: (data: UpdateData) => updateMyFeature(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['my-feature', id] })
    },
  })

  return {
    data: query.data,
    isLoading: query.isLoading,
    error: query.error,
    update: updateMutation.mutate,
    isUpdating: updateMutation.isPending,
  }
}
```

### Store Pattern (Optional)

```typescript
// features/my-feature/stores/my-feature-store.ts
import { create } from 'zustand'
import type { MyFeatureState } from '../types'

interface MyFeatureStore extends MyFeatureState {
  setData: (data: MyFeatureState) => void
  reset: () => void
}

const initialState: MyFeatureState = {
  items: [],
  selectedId: null,
}

export const useMyFeatureStore = create<MyFeatureStore>((set) => ({
  ...initialState,
  setData: (data) => set(data),
  reset: () => set(initialState),
}))
```

## Testing Pattern

```typescript
// features/my-feature/__tests__/my-component.test.tsx
import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MyComponent } from '../components/my-component'

describe('MyComponent', () => {
  it('renders successfully', () => {
    render(<MyComponent id="123" />)
    expect(screen.getByText('Expected Text')).toBeInTheDocument()
  })

  it('handles loading state', () => {
    render(<MyComponent id="123" />)
    expect(screen.getByText('Loading...')).toBeInTheDocument()
  })
})
```

## Development Commands

From `web/package.json:5-16`:

```bash
cd web

# Install dependencies
pnpm install

# Development
pnpm dev                   # Start dev server (Next.js with Turbopack)

# Build
pnpm build                 # Build for production
pnpm start                 # Start production server

# Code quality
pnpm lint                  # Next.js linter
pnpm format                # Format with Prettier
pnpm format:check          # Check formatting

# Testing
pnpm test                  # Run tests (Vitest)
pnpm test:watch            # Watch mode
pnpm test:ui               # Vitest UI
pnpm test:coverage         # Coverage report
```

## Best Practices

### 1. Import from Feature Index Only
```typescript
// ✅ Correct
import { useAuth } from '@/features/authentication'

// ❌ Wrong
import { useAuth } from '@/features/authentication/hooks/use-auth'
```

### 2. Keep Routes Thin
```typescript
// app/(dashboard)/projects/page.tsx
import { ProjectsList } from '@/features/projects'

export default function ProjectsPage() {
  return <ProjectsList />  // Delegate to feature component
}
```

### 3. Use Server Components by Default
Only add `'use client'` when absolutely necessary.

### 4. Type Everything
No `any` types. Use TypeScript strict mode.

### 5. Error Boundaries
Add error boundaries per feature and route:

```typescript
// app/(dashboard)/error.tsx
'use client'

export default function Error({ error, reset }: {
  error: Error
  reset: () => void
}) {
  return (
    <div>
      <h2>Something went wrong!</h2>
      <button onClick={reset}>Try again</button>
    </div>
  )
}
```

### 6. Loading States
Add `loading.tsx` for async routes:

```typescript
// app/(dashboard)/projects/loading.tsx
export default function Loading() {
  return <div>Loading projects...</div>
}
```

### 7. Test Critical Paths
Focus on:
- Authentication flows
- Navigation
- API calls
- User interactions
- Error states

### 8. shadcn/ui Integration
Use MCP tools to add components:

```bash
# Use shadcn MCP tools
mcp__shadcn__search_items_in_registries
mcp__shadcn__get_add_command_for_items
```

## Quick Decision Tree

**Creating new functionality?**
1. Is it feature-specific? → Create in `features/[feature]/`
2. Is it shared across features? → Create in `components/shared/`
3. Is it a UI primitive? → Use shadcn/ui or create in `components/ui/`
4. Is it utility logic? → Create in `lib/` or `hooks/`

**Adding state?**
1. Server data? → Use React Query in feature hook
2. Client-only state? → Use Zustand store in feature
3. Form state? → Use React Hook Form + Zod
4. URL state? → Use useSearchParams()

**Component decisions?**
1. Needs interactivity? → `'use client'`
2. Static content? → Server Component (default)
3. SEO important? → Server Component
4. Browser APIs needed? → `'use client'`

## Documentation

- **Architecture**: `web/ARCHITECTURE.md`
- **Backend API**: `CLAUDE.md` (API routes section)
- **Existing Features**: `web/src/features/` for reference patterns
