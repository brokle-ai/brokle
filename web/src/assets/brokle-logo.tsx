import { LucideProps } from 'lucide-react'

export const BrokleLogo = ({ ...props }: LucideProps) => (
  <svg
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    strokeLinecap="round"
    strokeLinejoin="round"
    {...props}
  >
    <circle cx="12" cy="12" r="10" />
    <path d="m8 12 2 2 4-4" />
  </svg>
)