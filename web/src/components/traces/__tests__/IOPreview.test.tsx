/**
 * Tests for IOPreview component
 */

import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { IOPreview } from '../IOPreview';

describe('IOPreview Component', () => {
  describe('ChatML Messages Rendering', () => {
    it('should render ChatML messages with chat UI', () => {
      const messages = [
        { role: 'user', content: 'Hello' },
        { role: 'assistant', content: 'Hi there!' },
      ];
      const value = JSON.stringify(messages);

      render(<IOPreview value={value} mimeType="application/json" label="Input" />);

      // Should show role badges
      expect(screen.getByText('user')).toBeInTheDocument();
      expect(screen.getByText('assistant')).toBeInTheDocument();

      // Should show content
      expect(screen.getByText('Hello')).toBeInTheDocument();
      expect(screen.getByText('Hi there!')).toBeInTheDocument();
    });

    it('should render tool calls in messages', () => {
      const messages = [
        {
          role: 'assistant',
          content: 'Using tool',
          tool_calls: [
            {
              id: 'call_1',
              type: 'function' as const,
              function: {
                name: 'get_weather',
                arguments: '{"location":"Bangalore"}',
              },
            },
          ],
        },
      ];
      const value = JSON.stringify(messages);

      render(<IOPreview value={value} mimeType="application/json" label="Output" />);

      expect(screen.getByText('Tool: get_weather')).toBeInTheDocument();
      expect(screen.getByText('{"location":"Bangalore"}')).toBeInTheDocument();
    });

    it('should render system messages', () => {
      const messages = [
        { role: 'system', content: 'You are a helpful assistant' },
        { role: 'user', content: 'Hello' },
      ];
      const value = JSON.stringify(messages);

      render(<IOPreview value={value} mimeType="application/json" label="Input" />);

      expect(screen.getByText('system')).toBeInTheDocument();
      expect(screen.getByText('You are a helpful assistant')).toBeInTheDocument();
    });
  });

  describe('Generic JSON Rendering', () => {
    it('should render generic JSON with JSON viewer', () => {
      const data = { endpoint: '/weather', query: 'Bangalore' };
      const value = JSON.stringify(data);

      render(<IOPreview value={value} mimeType="application/json" label="Input" />);

      // Should show formatted JSON
      expect(screen.getByText(/"endpoint"/)).toBeInTheDocument();
      expect(screen.getByText(/"query"/)).toBeInTheDocument();
    });

    it('should render nested JSON objects', () => {
      const data = {
        request: {
          endpoint: '/api/weather',
          params: { location: 'Bangalore' },
        },
      };
      const value = JSON.stringify(data);

      render(<IOPreview value={value} mimeType="application/json" label="Output" />);

      // Should show pretty-printed JSON
      const pre = screen.getByRole('code');
      expect(pre).toBeInTheDocument();
    });
  });

  describe('Plain Text Rendering', () => {
    it('should render plain text with text viewer', () => {
      const text = 'Hello world';

      render(<IOPreview value={text} mimeType="text/plain" label="Input" />);

      expect(screen.getByText('Hello world')).toBeInTheDocument();
    });

    it('should preserve whitespace in text', () => {
      const text = 'Line 1\nLine 2\n  Indented';

      render(<IOPreview value={text} mimeType="text/plain" label="Output" />);

      const textDiv = screen.getByText(/Line 1/);
      expect(textDiv).toHaveClass('whitespace-pre-wrap');
    });
  });

  describe('Truncation Handling', () => {
    it('should show truncation warning when truncated=true', () => {
      const value = 'x'.repeat(100);

      render(
        <IOPreview
          value={value}
          mimeType="text/plain"
          label="Input"
          truncated={true}
        />
      );

      expect(screen.getByText(/Content truncated/)).toBeInTheDocument();
      expect(screen.getByText(/exceeded 1MB limit/)).toBeInTheDocument();
    });
  });

  describe('Error Handling', () => {
    it('should fallback to text viewer on invalid JSON', () => {
      const invalidJSON = '{invalid json}';

      render(<IOPreview value={invalidJSON} mimeType="application/json" label="Input" />);

      // Should show error message
      expect(screen.getByText(/Invalid JSON format/)).toBeInTheDocument();

      // Should still display the raw value
      expect(screen.getByText('{invalid json}')).toBeInTheDocument();
    });

    it('should handle null value gracefully', () => {
      render(<IOPreview value={null} mimeType="application/json" label="Input" />);

      expect(screen.getByText(/No input data/)).toBeInTheDocument();
    });

    it('should handle undefined value gracefully', () => {
      render(<IOPreview value={undefined} mimeType="application/json" label="Output" />);

      expect(screen.getByText(/No output data/)).toBeInTheDocument();
    });
  });

  describe('MIME Type Auto-Detection Fallback', () => {
    it('should handle missing MIME type and auto-detect JSON', () => {
      const data = { key: 'value' };
      const value = JSON.stringify(data);

      // No mimeType provided - should try JSON parsing
      render(<IOPreview value={value} label="Input" />);

      // Should successfully render as JSON
      expect(screen.getByText(/"key"/)).toBeInTheDocument();
    });

    it('should handle missing MIME type with plain text', () => {
      const value = 'Plain text content';

      // No mimeType - should parse and fail, fallback to text
      render(<IOPreview value={value} label="Input" />);

      expect(screen.getByText('Plain text content')).toBeInTheDocument();
    });
  });
});
