import $ from 'jquery'
import 'bootstrap'
import {} from 'jquery-ujs'
import Turbolinks from 'turbolinks'
import Vue from 'vue'
import VueLazyload from 'vue-lazyload'

import vueLazyLoad from './common/vueLazyLoad'

document.addEventListener('turbolinks:load', event => {
  $.ajaxSetup({
    headers: { 'X-CSRF-Token': $('meta[name="csrf-token"]').attr('content') },
  })

  WebFont.load({
    google: {
      families: ['Raleway'],
    },
  })

  Vue.use(VueLazyload)

  Vue.nextTick(() => {
    vueLazyLoad.refresh()
  })

  new Vue({
    el: 'body',
  })
})

Turbolinks.start()
