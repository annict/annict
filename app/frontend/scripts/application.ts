import 'd3'
import 'dropzone'
import 'select2'
import { start } from 'turbolinks'
import Vue from 'vue'
import VueLazyload from 'vue-lazyload'

import App from './App'

import Home from './page_components/Home'

document.addEventListener('turbolinks:load', () => {
  Vue.use(VueLazyload)

  Vue.component('pc-home', Home)

  new App({ el: '.p-application' })
})

start()
