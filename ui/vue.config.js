const TerserPlugin = require("terser-webpack-plugin");
var path = require('path');
// const WebpackBundleAnalyzerPlugin = require('webpack-bundle-analyzer').BundleAnalyzerPlugin
const CompressionWebpackPlugin = require('compression-webpack-plugin');
const CopyWebPackPlugin = require('copy-webpack-plugin');
const productionGzipExtensions = ["js", "css"];
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
