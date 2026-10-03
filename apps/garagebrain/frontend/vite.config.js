import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// Backend for the /api proxy; the VPS sets it to the WireGuard address.
const apiTarget = process.env.API_TARGET || 'http://localhost:3002'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: apiTarget,
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
        target: apiTarget,
        changeOrigin: true,
      },
    },
  },
})
