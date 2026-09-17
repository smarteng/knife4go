const TerserPlugin = require("terser-webpack-plugin");
var path = require('path');
// const WebpackBundleAnalyzerPlugin = require('webpack-bundle-analyzer').BundleAnalyzerPlugin
const CompressionWebpackPlugin = require('compression-webpack-plugin');
const CopyWebPackPlugin = require('copy-webpack-plugin');
const productionGzipExtensions = ["js", "css"];

// 后端服务地址取自环境文件 .env.development 的 VUE_APP_PROXY_TARGET
// (由 @vue/cli-service 通过 dotenv 注入 process.env)，
// 未配置时回退到默认端口 14010，避免代理目标为空导致 dev server 启动失败。
const backendTarget = process.env.VUE_APP_PROXY_TARGET || "http://localhost:14010";

module.exports = {
  transpileDependencies: [
    /[/\\]node_modules[/\\](.+?)?mermaid(.*)/
  ],
  publicPath: ".",
  assetsDir: "webjars",
  outputDir: "dist",
  lintOnSave: false,
  productionSourceMap: false,
  indexPath: "doc.html",
  css: {
    loaderOptions: {
      less: {
        javascriptEnabled: true
      }
    }
  },
  devServer: {
    watchOptions: {
      ignored: /node_modules/
    },
    proxy: {
      // knife4j 前端在开发环境（doc.html 位于站点根路径）会请求相对路径
      // v3/api-docs/swagger-config，浏览器解析后打到 /v3/api-docs/swagger-config，
      // 而后端 knife4go 实际把该端点注册在 /swagger/v3/api-docs/swagger-config（受
      // uiPrefix 影响）。因此代理时补上 /swagger 前缀，让开发环境请求能正确落到后端。
      // 注意两点：
      //   1. webpack-dev-server 基于 http-proxy-middleware，重写规则必须用 pathRewrite；
      //      vite 的 rewrite 选项在 webpack 下会被忽略。
      //   2. vue-cli 的 prepareProxy 把每个 key 当正则交给 String.match 做「子串匹配」，
      //      而非前缀匹配。若写成 '/swagger'，/v3/api-docs/swagger-config 也含子串
      //      "/swagger" 会被它抢先命中，从而绕过本规则。故 context 一律用 '^/xxx' 锚定开头。
      // 生产环境后端 docPath 通常为 /swagger/index.html，浏览器会自动带上 /swagger
      // 前缀，不再走此规则，故本规则仅对开发环境生效。
      '^/v3/api-docs': {
        target: backendTarget,
        changeOrigin: true,
        pathRewrite: { '^/v3/api-docs': '/swagger/v3/api-docs' }
      },
      // 后端 knife4go 把文档 JSON 注册在 /swagger/doc.json，/swagger 前缀原样透传。
      '^/swagger': {
        target: backendTarget,
        changeOrigin: true
      }
    }
  },
  configureWebpack: {
    optimization: {
      minimizer: [
        new TerserPlugin({
          terserOptions: {
            ecma: undefined,
            warnings: false,
            parse: {},
            compress: {
              drop_console: true,
              drop_debugger: true,
              pure_funcs: ['console.log', 'console.debug', 'window.console.log', 'window.console.debug'] // 移除console
            }
          },
        }),

      ]
    },
    plugins: [
      new CompressionWebpackPlugin({
        algorithm: "gzip",
        test: new RegExp("\\.(" + productionGzipExtensions.join("|") + ")$"),
        threshold: 10240,
        minRatio: 0.8
      }),
      new CopyWebPackPlugin([
        { from: path.resolve(__dirname, 'public/oauth'), to: path.resolve(__dirname, 'dist/webjars/oauth') }
      ])
    ]
  }
};
