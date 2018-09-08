import 'd3'
import 'dropzone'
import 'select2'
import Turbolinks from 'turbolinks'
import Vue from 'vue'
import VueLazyload from 'vue-lazyload'

import App from './App'

import Home from './components/Home'

document.addEventListener('turbolinks:load', async () => {
  Vue.use(VueLazyload)

  Vue.component('ann-home', Home)

  new App({ el: '.p-application' })
})

Turbolinks.start()
