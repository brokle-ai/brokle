"use client";

import Image from "next/image";

interface BrokleLogoProps {
  variant?: "icon" | "full" | "stacked";
  size?: "xs" | "sm" | "md" | "lg" | "xl";
  className?: string;
  forceTheme?: "light" | "dark";
}

const sizeMap = {
  xs: { icon: 16, full: { width: 80, height: 20 }, stacked: { width: 40, height: 30 } },
  sm: { icon: 24, full: { width: 120, height: 30 }, stacked: { width: 60, height: 45 } },
  md: { icon: 32, full: { width: 160, height: 40 }, stacked: { width: 80, height: 60 } },
  lg: { icon: 48, full: { width: 200, height: 50 }, stacked: { width: 100, height: 75 } },
  xl: { icon: 64, full: { width: 240, height: 60 }, stacked: { width: 120, height: 90 } },
};

export function BrokleLogo({
  variant = "full",
  size = "md",
  className = "",
  forceTheme,
}: BrokleLogoProps) {
  // Get dimensions based on variant and size
  const dimensions = sizeMap[size];

  // Get source files for light and dark themes
  const getSources = () => {
    switch (variant) {
      case "icon":
        return { light: "/logo/icon.svg", dark: "/logo/icon-white.svg" };
      case "stacked":
        return { light: "/logo/logo-stacked.svg", dark: "/logo/logo-stacked-white.svg" };
      case "full":
      default:
        return { light: "/logo/logo-full.svg", dark: "/logo/logo-full-white.svg" };
    }
  };

  // Get dimensions for the variant
  const getDimensions = () => {
    switch (variant) {
      case "icon":
        return { width: dimensions.icon, height: dimensions.icon };
      case "stacked":
        return dimensions.stacked;
      case "full":
      default:
        return dimensions.full;
    }
  };

  const { width, height } = getDimensions();
  const sources = getSources();

  // If forceTheme is set, render only that theme's logo
  if (forceTheme) {
    return (
      <Image
        src={forceTheme === "dark" ? sources.dark : sources.light}
        alt="Brokle"
        width={width}
        height={height}
        className={className}
        priority
      />
    );
  }

  // CSS-based theme switching: render both images, CSS toggles visibility
  // This prevents hydration mismatch and FOUC (Flash of Unstyled Content)
  return (
    <span className={`inline-flex ${className}`}>
      {/* Light theme logo - visible by default, hidden in dark mode */}
      <Image
        src={sources.light}
        alt="Brokle"
        width={width}
        height={height}
        className="block dark:hidden"
        priority
      />
      {/* Dark theme logo - hidden by default, visible in dark mode */}
      <Image
        src={sources.dark}
        alt="Brokle"
        width={width}
        height={height}
        className="hidden dark:block"
        priority
      />
    </span>
  );
}

export default BrokleLogo;
