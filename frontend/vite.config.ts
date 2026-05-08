import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import { fileURLToPath, URL } from 'node:url'
import path from 'path'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    tailwindcss(), // 【关键修复】激活 Tailwind CSS v4 插件
  ],
  // --- 新增 build 配置 ---
  build: {
    // 1. 设置输出目录为同级 backend 目录下的 static 文件夹
    outDir: path.resolve(__dirname, '../backend/static'),

    // 2. 必须设置，因为输出目录在项目根目录之外，Vite 默认不会清空它
    emptyOutDir: true,
  },
  resolve: {
    alias: {
      // 确保 App.vue 中的 '@/components/...' 能够被正确解析
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:8080',
        changeOrigin: true,
      }
    }
  }
})