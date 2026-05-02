interface BrokleLogoProps {
  variant?: 'icon' | 'full' | 'stacked'
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl'
  className?: string
  forceTheme?: 'light' | 'dark'
}

const sizeMap = {
  xs: { icon: 16, full: { width: 80, height: 20 }, stacked: { width: 40, height: 30 } },
  sm: { icon: 24, full: { width: 120, height: 30 }, stacked: { width: 60, height: 45 } },
  md: { icon: 32, full: { width: 160, height: 40 }, stacked: { width: 80, height: 60 } },
  lg: { icon: 48, full: { width: 200, height: 50 }, stacked: { width: 100, height: 75 } },
  xl: { icon: 64, full: { width: 240, height: 60 }, stacked: { width: 120, height: 90 } },
}

export function BrokleLogo({
  variant = 'full',
  size = 'md',
  className = '',
  forceTheme,
}: BrokleLogoProps) {
  const dimensions = sizeMap[size]

  const sources =
    variant === 'icon'
      ? { light: '/logo/icon.svg', dark: '/logo/icon-white.svg' }
      : variant === 'stacked'
        ? { light: '/logo/logo-stacked.svg', dark: '/logo/logo-stacked-white.svg' }
        : { light: '/logo/logo-full.svg', dark: '/logo/logo-full-white.svg' }

  const { width, height } =
    variant === 'icon'
      ? { width: dimensions.icon, height: dimensions.icon }
      : variant === 'stacked'
        ? dimensions.stacked
        : dimensions.full

  if (forceTheme) {
    return (
      <img
        src={forceTheme === 'dark' ? sources.dark : sources.light}
        alt="Brokle"
        width={width}
        height={height}
        className={className}
      />
    )
  }

  return (
    <span className={`inline-flex ${className}`}>
      <img
        src={sources.light}
        alt="Brokle"
        width={width}
        height={height}
        className="block dark:hidden"
      />
      <img
        src={sources.dark}
        alt="Brokle"
        width={width}
        height={height}
        className="hidden dark:block"
      />
    </span>
  )
}

export default BrokleLogo
