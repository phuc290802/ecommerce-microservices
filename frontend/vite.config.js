import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

const gatewayTarget = process.env.GATEWAY_URL || 'http://localhost:8080'

export default defineConfig({
  plugins: [vue()],
  server: {
    host: '0.0.0.0',
    port: 5173,
    proxy: {
      '/api': {
        target: gatewayTarget,
        changeOrigin: true,
        secure: false,
      }
    }
  }
})
