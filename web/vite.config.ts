import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  plugins: [react()],
  envDir: '..',
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        ws: true,
      },
      '/realms': {
        target: 'http://localhost:8088',
        changeOrigin: true,
      },
      '/resources': {
        target: 'http://localhost:8088',
        changeOrigin: true,
      },
    },
  },
})
