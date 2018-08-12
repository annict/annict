import $ from 'jquery'
import 'bootstrap'
import 'select2'
import 'd3'
import 'dropzone'
import {} from 'jquery-ujs'
import Turbolinks from 'turbolinks'
import Vue from 'vue'
import VueLazyload from 'vue-lazyload'
import VueWait from 'vue-wait'

import App from './v3/App'

import HomeGuest from './v3/components/HomeGuest'

document.addEventListener('turbolinks:load', async function() {
  Vue.use(VueLazyload)
  Vue.use(VueWait)

  Vue.component('ann-home-guest', HomeGuest)

  new Vue({
    wait: new VueWait(),

    el: '.p-application',

    async created() {
      this.$wait.start('Setup App')

      await App.setup()

      const isProduction = await App.isProduction()
      Vue.config.silent = isProduction
      Vue.config.devtools = !isProduction
      Vue.config.performance = !isProduction
      Vue.config.productionTip = !isProduction

      this.$wait.end('Setup App')
    },
  })
})

Turbolinks.start()
