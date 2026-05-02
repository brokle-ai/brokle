import { render, screen, waitFor } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { PromptEditor } from '../prompt-editor'

describe('PromptEditor', () => {
  it('renders the value', async () => {
    render(<PromptEditor value="hello world" ariaLabel="demo" />)
    await waitFor(() => {
      expect(screen.getByText(/hello world/)).toBeInTheDocument()
    })
  })

  it('applies read-only mode without the text being editable', async () => {
    const { container } = render(
      <PromptEditor value="{{ greeting }}" language="jinja" readOnly ariaLabel="ro" />,
    )
    await waitFor(() => {
      expect(screen.getByText(/greeting/)).toBeInTheDocument()
    })
    // CM6 mounts contenteditable on .cm-content inside the editor wrapper.
    const content = container.querySelector('.cm-content')
    expect(content).toHaveAttribute('contenteditable', 'false')
  })
})
