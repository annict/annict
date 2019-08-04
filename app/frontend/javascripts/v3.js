import 'bootstrap'
import Turbolinks from 'turbolinks'
import Vue from 'vue'
import VueI18n from 'vue-i18n'

import { fetchViewerQuery } from './v3/queries'

import WorkDetail from './v3/components/pages/WorkDetail.vue'

import messages from './v3/messages'

document.addEventListener('turbolinks:load', _event => {
  Vue.use(VueI18n)

  const i18n = new VueI18n({
    locale: 'ja',
    fallbackLocale: 'en',
    messages,
  })

  Vue.component('c-work-detail', WorkDetail)

  new Vue({
    i18n,
    el: '#app',
    data() {
      return {
        viewer: null,
      }
    },

    async created() {
      const result = await fetchViewerQuery()
      this.viewer = result.data.viewer
    },

    methods: {
      isSignedIn() {
        return !!this.viewer
      },
    },
  })
})

Turbolinks.start()
