import { createApp } from 'vue'
import './style/knife4j.less'
// 显式引入 ant-design-vue 函数式 API 的样式（unplugin-vue-components 只处理模板标签
// 形式的按需组件，对 import { message, Modal, notification } from 'ant-design-vue' 这类
// JS 引用不会自动注入 CSS，必须手动引入，否则弹出提示不可见）
import 'ant-design-vue/es/message/style/css'
import 'ant-design-vue/es/notification/style/css'
import 'ant-design-vue/es/modal/style/css'
import App from './App.vue'
import { setupStore } from './store/index.js'
import router from '@/router/index.js'
import { setupI18n } from '@/lang/index.js'
import { createFromIconfontCN } from '@ant-design/icons-vue'

String.prototype.gblen = function () {
  let len = 0
  for (let i = 0; i < this.length; i++) {
    if (this.charCodeAt(i) > 127 || this.charCodeAt(i) == 94) {
      len += 2;
    } else {
      len++;
    }
  }
  return len;
}

String.prototype.startWith = function (str) {
  const reg = new RegExp("^" + str)
  return reg.test(this);
}

/***
 * 自定义图标
 */
import iconFront from './assets/iconfonts/iconfont.js'
const MyIcon = createFromIconfontCN({
  scriptUrl: iconFront
})

const app = createApp(App)
app.use(router)
app.component('my-icon', MyIcon)
setupStore(app)
setupI18n(app)
app.mount('#app')
