/**
 * User utility functions for displaying user information
 */

interface GetUserInitialsInput {
  firstName?: string
  lastName?: string
  name?: string // Full name to be split if firstName/lastName not provided
  email?: string
}

/**
 * Generate user initials from name or email
 *
 * Priority order:
 * 1. firstName + lastName → first letter of each (uppercase)
 * 2. name (full name) → split and use first letter of first/last word
 * 3. firstName only → first 2 letters
 * 4. lastName only → first 2 letters
 * 5. email → first 2 letters before @
 * 6. Fallback → "?"
 *
 * @param input Object containing firstName, lastName, name, or email
 * @returns User initials (1-2 uppercase letters)
 *
 * @example
 * getUserInitials({ firstName: 'John', lastName: 'Doe' }) // "JD"
 * getUserInitials({ name: 'John Doe' }) // "JD"
 * getUserInitials({ firstName: 'John' }) // "JO"
 * getUserInitials({ email: 'john@example.com' }) // "JO"
 * getUserInitials({}) // "?"
 */
export function getUserInitials(input: GetUserInitialsInput): string {
  const { firstName, lastName, name, email } = input

  // If full name provided without firstName/lastName, split it
  if (name && !firstName && !lastName) {
    const nameParts = name.trim().split(' ').filter(Boolean)
    if (nameParts.length >= 2) {
      // Use first and last part
      return getUserInitials({
        firstName: nameParts[0],
        lastName: nameParts[nameParts.length - 1],
        email
      })
    } else if (nameParts.length === 1) {
      // Single name, use first 2 letters
      return nameParts[0].substring(0, 2).toUpperCase()
    }
  }

  // firstName + lastName
  if (firstName && lastName) {
    return `${firstName[0]}${lastName[0]}`.toUpperCase()
  }

  // firstName only
  if (firstName) {
    return firstName.substring(0, 2).toUpperCase()
  }

  // lastName only
  if (lastName) {
    return lastName.substring(0, 2).toUpperCase()
  }

  // email fallback
  if (email) {
    const emailPrefix = email.split('@')[0]
    if (emailPrefix) {
      return emailPrefix.substring(0, 2).toUpperCase()
    }
  }

  // Ultimate fallback
  return '?'
}
