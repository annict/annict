import 'bootstrap'
import Turbolinks from 'turbolinks'
import Vue from 'vue'

import WorkDetail from './v3/components/pages/WorkDetail.vue'

document.addEventListener('turbolinks:load', _event => {
  Vue.component('c-work-detail', WorkDetail)

  new Vue({
    el: '#app',
  })
})

Turbolinks.start()
