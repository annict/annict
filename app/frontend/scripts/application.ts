import 'd3'
import 'dropzone'
import 'select2'
import { start } from 'turbolinks'
import Vue from 'vue'
import VueApollo from 'vue-apollo'
import VueLazyload from 'vue-lazyload'

import App from './App'
import Apollo from './utils/Apollo'

document.addEventListener('turbolinks:load', () => {
  Vue.use(VueApollo)
  Vue.use(VueLazyload)

  new App({
    el: '.c-app',
    provide: Apollo.provider(),
  })
})

start()
