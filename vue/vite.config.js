import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueJsx from '@vitejs/plugin-vue-jsx'
import Components from 'unplugin-vue-components/vite'
import { AntDesignVueResolver } from 'unplugin-vue-components/resolvers'
import viteCompression from 'vite-plugin-compression';
import removeConsole from 'vite-plugin-remove-console';
import { resolve } from 'path'
import { nodePolyfills } from 'vite-plugin-node-polyfills'

// https://vitejs.dev/config/
export default defineConfig({
  base: './',
  plugins: [
    vue(),
    vueJsx(),
    Components({
      resolvers: [AntDesignVueResolver()]
    }),
    nodePolyfills(),
    viteCompression({
      deleteOriginFile: false, //删除源文件
      threshold: 10240, //压缩前最小文件大小
      algorithm: 'gzip', //压缩算法
      ext: '.gz', //文件类型
    }),
    // removeConsole()
  ],
  resolve: {
    alias: [
      { find: '@', replacement: resolve(__dirname, 'src') },
      { find: /^~/, replacement: '' },
    ]
  },
  // 开启less支持
  css: {
    preprocessorOptions: {
      less: {
        javascriptEnabled: true
      }
    }
  },
  server: {
    host: true,
    proxy: {
      '/swagger': {
        target: `http://localhost:14010`,
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/swagger/, '')
      }
    }
  },
  build: {
    // 提高 chunk 大小警告阈值（单位 KB），避免第三方大依赖触发提示
    chunkSizeWarningLimit: 1500,
    rollupOptions: {
      input: 'doc.html',
      output: {
        chunkFileNames: 'webjars/js/[name]-[hash].js',
        entryFileNames: 'webjars/js/[name]-[hash].js',
        assetFileNames: 'webjars/[ext]/[name]-[hash].[ext]',
        // 手动拆分 vendor：把重型第三方依赖拆到独立 chunk，浏览器可并行下载 + 长期缓存
        manualChunks: {
          'vendor-vue': ['vue', 'vue-router', 'vue-i18n', 'pinia'],
          'vendor-antd': ['ant-design-vue', '@ant-design/icons-vue'],
          'vendor-editor': ['vue3-ace-editor', 'ace-builds'],
          'vendor-mermaid': ['mermaid'],
          'vendor-utils': [
            'lodash',
            'axios',
            'dayjs',
            'marked',
            'crypto-js',
            'qs',
            'json5',
            'xml2js',
            'clipboard',
            'async',
            'localforage'
          ]
        }
      }
    }
  }
})
