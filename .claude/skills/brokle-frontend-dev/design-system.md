# Brokle Design System

Reference for colors, typography, spacing, and visual design patterns.

## Brand Identity

Brokle is an **AI observability platform**. Design reflects:
- **Professional & Data-Focused**: Clean, technical aesthetic
- **Trustworthy & Reliable**: Enterprise-grade quality
- **Modern & Innovative**: Forward-thinking without being flashy

## Color System

Uses **OKLCH color space** with CSS variables for dark mode support.

### Semantic Color Classes

**Always use CSS variables, never hardcode colors:**

```tsx
// Correct
<div className="bg-primary text-primary-foreground">CTA</div>
<div className="bg-accent text-accent-foreground">Accent</div>
<div className="text-destructive">Error</div>
<div className="bg-muted text-muted-foreground">Subtle</div>

// Wrong
<div className="bg-blue-600 text-white">Button</div>
```

| Class | Use For |
|-------|---------|
| `bg-primary` / `text-primary` | Primary brand actions |
| `bg-secondary` / `text-secondary` | Secondary actions |
| `bg-accent` / `text-accent` | Accent elements |
| `bg-destructive` / `text-destructive` | Errors, delete actions |
| `bg-muted` / `text-muted-foreground` | Subtle backgrounds, secondary text |
| `border-border` | All borders |
| `ring-ring` | Focus rings |

### Chart Colors

```tsx
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

## Dark Mode

Uses `next-themes` with automatic CSS variable switching.

```tsx
'use client'
import { useTheme } from 'next-themes'

export function ThemeToggle() {
  const { theme, setTheme } = useTheme()
  return (
    <Button onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}>
      Toggle
    </Button>
  )
}
```

**Principles:**
1. Use CSS variables - colors switch automatically
2. Test both modes - verify contrast in light and dark
3. Maintain 4.5:1 contrast ratio for accessibility

## Typography

### Font Stack

Currently uses **Inter** (loaded in `app/layout.tsx`).

```tsx
// All UI text uses Inter via font-sans
<h1 className="font-sans text-4xl font-bold">Heading</h1>
<code className="font-mono text-sm">Code</code>
```

### Scale

| Class | Size | Use |
|-------|------|-----|
| `text-xs` | 12px | Labels, captions |
| `text-sm` | 14px | Secondary text |
| `text-base` | 16px | Body text |
| `text-lg` | 18px | Emphasis |
| `text-xl` | 20px | Subheadings |
| `text-2xl` | 24px | H3 |
| `text-3xl` | 30px | H2 |
| `text-4xl` | 36px | H1 |
| `text-5xl` | 48px | Display/Hero |

### Weight Hierarchy

Use high contrast between weights:
- **Display**: `font-bold` (700) or `font-extrabold` (800)
- **Headings**: `font-semibold` (600) or `font-bold` (700)
- **Body**: `font-normal` (400) or `font-medium` (500)

## Spacing & Layout

| Element | Classes |
|---------|---------|
| Container max | `max-w-[1280px]` (xl) |
| Content padding | `px-4 md:px-6 lg:px-8` |
| Section spacing | `py-12 md:py-16 lg:py-24` |
| Card padding | `p-6 md:p-8` |
| Grid gaps | `gap-4 md:gap-6 lg:gap-8` |

## Depth & Shadows

| Class | Use |
|-------|-----|
| `shadow-sm` | Cards on white background |
| `shadow-md` | Dropdowns, popovers |
| `shadow-lg` | Modals, dialogs |
| `shadow-xl` | Tooltips, notifications |

```tsx
// Overlapping elements for depth
<div className="relative">
  <div className="absolute -top-4 -left-4 w-24 h-24 bg-blue-500/10 rounded-full blur-2xl" />
  <Card className="relative z-10">Content</Card>
</div>
```

## Background Patterns

Avoid plain white/solid colors. Use:

```tsx
// Hero - atmospheric gradient
<section className="relative bg-gradient-to-br from-neutral-900 via-primary-900 to-neutral-900">
  <div className="absolute inset-0 bg-[url('/grid.svg')] opacity-20" />
</section>

// Dashboard - subtle depth
<section className="bg-gradient-to-b from-neutral-50 to-white">
  <div className="absolute inset-0 bg-[url('/dots.svg')] opacity-5" />
</section>
```

## Responsive Breakpoints

| Prefix | Width | Device |
|--------|-------|--------|
| `sm:` | 640px | Small tablets |
| `md:` | 768px | Tablets |
| `lg:` | 1024px | Laptops |
| `xl:` | 1280px | Desktops |
| `2xl:` | 1536px | Large desktops |

```tsx
// Stack mobile, grid desktop
<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">

// Hide mobile, show desktop
<div className="hidden lg:block">Desktop only</div>

// Responsive text
<h1 className="text-3xl md:text-4xl lg:text-5xl font-bold">
```
