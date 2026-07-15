import { fileURLToPath, URL } from 'node:url'
import { writeFileSync } from 'node:fs'

import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vite'

export default defineConfig({
  plugins: [
    vue(),
    {
      name: 'preserve-dist-placeholder',
      closeBundle() {
        writeFileSync(fileURLToPath(new URL('./dist/.gitkeep', import.meta.url)), '')
      }
    }
  ],
  build: {
    emptyOutDir: true
  },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    port: 5173,
    strictPort: true
  }
})
