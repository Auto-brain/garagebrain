import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:3002',
        changeOrigin: true,
      },
    },
  },
  // `vite preview` отдаёт собранный dist/. host:true биндит на 0.0.0.0 (доступ
  // по внешнему IP), а /api проксируется на backend — так SPA остаётся
  // same-origin и не упирается в CORS.
  preview: {
    host: true,
    port: 4173,
    proxy: {
      '/api': {
        target: 'http://localhost:3002',
        changeOrigin: true,
      },
    },
  },
})
