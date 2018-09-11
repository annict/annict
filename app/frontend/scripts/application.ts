import 'd3'
import 'dropzone'
import 'select2'
import { start } from 'turbolinks'
import Vue from 'vue'
import VueLazyload from 'vue-lazyload'

import App from './App'

document.addEventListener('turbolinks:load', () => {
  Vue.use(VueLazyload)

  new App({ el: '.c-app' })
})

start()
