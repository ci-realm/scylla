// The Vue build version to load with the `import` command
// (runtime-only or standalone) has been set in webpack.base.conf with an alias.
import Vue from 'vue'
import VueMomentLib from 'vue-moment-lib'
import VueNativeSock from 'vue-native-websocket'
import VueHighlightJS from 'vue-highlightjs'
import VueTimeago from 'vue-timeago'
import dateLocaleDe from 'date-fns/locale/de'
import Vuetify from 'vuetify/lib'
import 'vuetify/src/stylus/app.styl'
import 'vuetify/dist/vuetify.min.css'
import '@mdi/font/css/materialdesignicons.css'
import 'highlight.js/styles/gruvbox-dark.css'

import App from './App'
import store from './store'
import router from './router'

Vue.config.productionTip = false
// Vue.use(ElementUI)
// Vue.use(AtComponents)
Vue.use(Vuetify, {
  iconfont: 'mdi',
  theme: {
    primary: '#b2dfdb',
    secondary: '#424242',
    accent: '#82B1FF',
    error: '#FF5252',
    info: '#2196F3',
    success: '#4CAF50',
    warning: '#FFC107',
  },
})
Vue.use(VueHighlightJS)
Vue.use(VueNativeSock, 'ws://localhost:7100/socket', {
  store,
  format: 'json',
  reconnection: true,
})
Vue.use(VueTimeago, {
  name: 'Timeago', // Component name, `Timeago` by default
  locale: 'en', // Default locale
  // We use `date-fns` under the hood
  // So you can use all locales from it
  locales: {
    // 'zh-CN': require('date-fns/locale/zh_cn'),
    // 'ja': require('date-fns/locale/ja'),
    de: dateLocaleDe,
  },
})
Vue.use(VueMomentLib)

/* eslint-disable no-new */
new Vue({
  el: '#app',
  router,
  store,
  template: '<App/>',
  components: {
    App,
  },
})
