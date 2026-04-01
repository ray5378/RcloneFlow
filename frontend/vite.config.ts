import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

export default defineConfig({
  plugins: [vue()],
  root: '.',
  base: './',
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  build: {
    outDir: '../web',
    emptyOutDir: true,
    rollupOptions: {
      output: {
        entryFileNames: 'assets/index-[name].js',
        chunkFileNames: 'assets/index-[name].js',
        assetFileNames: 'assets/index-[name].[ext]'
      }
    }
  },
  server: {
    port: 4200,
    proxy: {
      '/api': {
        target: 'http://localhost:17870',
        changeOrigin: true,
      },
    },
  },
})
