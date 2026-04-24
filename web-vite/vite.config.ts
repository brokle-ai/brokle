import { defineConfig, loadEnv } from 'vite'
import { TanStackRouterVite } from '@tanstack/router-plugin/vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import csp from 'vite-plugin-csp-guard'
import path from 'node:path'
import { cspPolicy } from './src/csp-policy'

// Vite 8 + Rolldown (default). `advancedChunks.groups` replaces Rollup's
// `manualChunks` — priority-based: higher wins on overlap.

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), 'VITE_')

  return {
    plugins: [
      TanStackRouterVite({ target: 'react', autoCodeSplitting: true }),
      react(),
      tailwindcss(),
      csp({
        dev: { run: false },
        build: { sri: true },
        policy: cspPolicy,
      }),
    ],
    resolve: {
      alias: {
        '@': path.resolve(__dirname, './src'),
      },
    },
    server: {
      port: 3001,
      strictPort: true,
      proxy: {
        '/api/v1': {
          target: env.VITE_PROXY_TARGET || 'http://localhost:8080',
          changeOrigin: false,
          secure: false,
        },
        '/v1': {
          target: env.VITE_PROXY_TARGET || 'http://localhost:8080',
          changeOrigin: false,
          secure: false,
        },
      },
    },
    build: {
      target: 'es2022',
      sourcemap: true,
      // Vite 8 + Rolldown: priority-based chunk groups. Let the bundler's
      // heuristics split the app code; we only pin framework-y node_modules
      // so each grows in isolation and browser cache is preserved across
      // deploys that touch only app code.
      rolldownOptions: {
        output: {
          codeSplitting: {
            groups: [
              {
                name: 'react',
                test: /[\\/]node_modules[\\/](react|react-dom|scheduler)[\\/]/,
                priority: 10,
              },
              {
                name: 'tanstack',
                test: /[\\/]node_modules[\\/]@tanstack[\\/]/,
                priority: 9,
              },
              {
                name: 'radix',
                test: /[\\/]node_modules[\\/]@radix-ui[\\/]/,
                priority: 8,
              },
              {
                name: 'codemirror',
                test: /[\\/]node_modules[\\/](@codemirror|@lezer)[\\/]/,
                priority: 7,
              },
            ],
          },
        },
      },
    },
    test: {
      environment: 'jsdom',
      globals: true,
      setupFiles: ['./src/test/setup.ts'],
    },
  }
})
