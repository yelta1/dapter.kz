import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

import { cloudflare } from "@cloudflare/vite-plugin";

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss(), cloudflare()],
  server: {
    port: 3000, // Порт для фронтенда
    proxy: {
      // Проксируем API запросы на бэкенд Go
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      // Проксируем загруженные чеки на бэкенд Go
      '/uploads': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})