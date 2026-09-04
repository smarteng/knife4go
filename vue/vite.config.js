import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueJsx from '@vitejs/plugin-vue-jsx'
import Components from 'unplugin-vue-components/vite'
import { AntDesignVueResolver } from 'unplugin-vue-components/resolvers'
import viteCompression from 'vite-plugin-compression';
import removeConsole from 'vite-plugin-remove-console';
import { resolve, dirname } from 'path'
import { fileURLToPath } from 'url'
import { nodePolyfills } from 'vite-plugin-node-polyfills'

// package.json 设置 "type": "module" 后 __dirname 在 ESM 下不再存在，
// 用 import.meta.url 手工推导等价值，供下方 resolve(__dirname, 'src') 使用。
const __dirname = dirname(fileURLToPath(import.meta.url))

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
      // 后端 knife4go 将 API 文档 JSON 端点注册在 /swagger/doc.json，
      // /swagger 前缀的请求原样透传到后端，不做 rewrite。
      '/swagger': {
        target: `http://localhost:14010`,
        changeOrigin: true
      },
      // knife4j 前端在开发环境（doc.html 位于站点根路径）会请求相对路径
      // v3/api-docs/swagger-config，浏览器解析后打到 /v3/api-docs/swagger-config，
      // 而后端 knife4go 实际把该端点注册在 /swagger/v3/api-docs/swagger-config（受
      // uiPrefix 影响）。因此代理时补上 /swagger 前缀，让开发环境请求能正确落到后端。
      // 生产环境后端 docPath 通常为 /swagger/index.html，浏览器会自动加上 /swagger
      // 前缀，不再走此规则，故本规则仅对开发环境生效。
      '/v3/api-docs': {
        target: `http://localhost:14010`,
        changeOrigin: true,
        rewrite: (p) => '/swagger' + p
      }
    }
  },
  build: {
    // 提高 chunk 大小警告阈值（单位 KB），避免第三方大依赖触发提示
    chunkSizeWarningLimit: 1500,
    rollupOptions: {
      input: 'doc.html',
      // 静默 ant-design-vue 3.x 内部 _vueuse 模块产生的 /* #__PURE__ */ 注释位置告警。
      // 该告警源于第三方库源码（node_modules/ant-design-vue/es/_util/hooks/_vueuse/*.js），
      // Rollup 5+ 对 pure 注释位置检查更严格，会自动移除位置不合法的注释以避免副作用，
      // 不影响构建结果和运行时行为。ant-design-vue 3.x 已停止维护，官方不会再修复；
      // 因此在此处按已知模式过滤，避免每次构建刷屏几十条噪音。
      // 其他 Rollup 告警仍会正常输出，以便发现真实问题。
      onwarn(warning, defaultHandler) {
        if (
          warning.code === 'INVALID_ANNOTATION' &&
          warning.message &&
          warning.message.includes('/* #__PURE__ */') &&
          warning.id &&
          warning.id.includes('/ant-design-vue/es/_util/hooks/_vueuse/')
        ) {
          return
        }
        defaultHandler(warning)
      },
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
