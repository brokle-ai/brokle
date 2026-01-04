# Animation Patterns

Motion and animation guidance for Brokle frontend. Uses CSS animations via `tailwindcss-animate` and custom keyframes in `globals.css`.

## Available Utility Classes

From `globals.css`:

| Class | Effect |
|-------|--------|
| `fade-in` | Fade in on mount |
| `slide-in-right` | Slide from right |
| `slide-in-left` | Slide from left |
| `zoom-in` | Zoom in smoothly |
| `btn-hover` | Button scale on hover |
| `card-hover` | Card lift on hover |
| `pulse-subtle` | Subtle pulsing |
| `bounce-gentle` | Gentle bouncing |

## Tailwind Built-in Animations

```tsx
<div className="animate-pulse">Loading...</div>
<Loader2 className="animate-spin" />
<div className="animate-bounce">Badge</div>
```

## Micro-Interactions

### Button States

```tsx
// shadcn Button already has active:scale-95
<Button className="hover:shadow-md transition-all">Click</Button>

// Enhanced hover
<Button className="hover:scale-105 transition-transform">Hover</Button>
```

### Card Hover

```tsx
// Use utility class
<Card className="card-hover">Lifts on hover</Card>

// Custom
<Card className="transition-all hover:shadow-lg hover:-translate-y-1">
  Custom lift
</Card>
```

### Loading Skeletons

```tsx
import { Skeleton } from '@/components/ui/skeleton'

<Card className="p-6">
  <div className="space-y-4">
    <Skeleton className="h-4 w-24" />
    <Skeleton className="h-8 w-32" />
    <Skeleton className="h-32 w-full" />
  </div>
</Card>
```

## Transitions

### Modal/Dialog Entry

shadcn Dialog has built-in animations:

```tsx
<DialogContent className="animate-in fade-in-0 zoom-in-95">
  {/* Content */}
</DialogContent>
```

### Toast Notifications

Sonner handles animations automatically:

```tsx
toast.success('Saved!') // Animates in from position
```

### Page Transitions

```tsx
export function DashboardPage() {
  return (
    <div className="fade-in">
      <h1>Dashboard</h1>
    </div>
  )
}
```

## Performance Guidelines

1. **Use CSS over JS** for hover, focus, simple state changes
2. **Support reduced motion**:

```css
@media (prefers-reduced-motion: reduce) {
  * {
    animation-duration: 0.01ms !important;
    transition-duration: 0.01ms !important;
  }
}
```

3. **Prioritize key interactions** - Don't add animations everywhere. Focus on:
   - Page loads
   - Data updates
   - User feedback moments

## Transition Classes

```tsx
// Smooth all transitions
<div className="transition-all duration-200">

// Specific properties
<div className="transition-colors duration-150">
<div className="transition-transform duration-200">
<div className="transition-opacity duration-300">

// With easing
<div className="transition-all ease-in-out">
<div className="transition-all ease-out">
```
