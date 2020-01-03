import ujs from '@rails/ujs'
import 'bootstrap'
import Turbolinks from 'turbolinks'
import Vue from 'vue'
import VueCompositionApi from '@vue/composition-api'
import VueI18n from 'vue-i18n'

import Track from "./v3/page-components/Track.vue";
import WorkDetail from './v3/page-components/WorkDetail.vue'

import { formatDateTime, formatDomain } from './v3/filters'
import messages from './v3/messages'
import { FetchViewerQuery } from './v3/queries'

Vue.config.productionTip = false
Vue.use(VueCompositionApi)
Vue.use(VueI18n)

const i18n = new VueI18n({
  locale: window.AnnConfig.isDomainJp ? 'ja' : 'en',
  messages,
})

Vue.filter('formatDateTime', formatDateTime)
Vue.filter('formatDomain', formatDomain)

Vue.component('pc-track', Track)
Vue.component('pc-work-detail', WorkDetail)

document.addEventListener('turbolinks:load', _event => {
  window.WebFont.load({
    google: {
      families: ['Raleway'],
    },
  })

  new Vue({
    i18n,
    el: '#app',
    data() {
      return {
        viewer: null,
      }
    },

    async created() {
      const result = await new FetchViewerQuery().execute()
      this.viewer = result.data.viewer
    },

    methods: {
      isSignedIn() {
        return !!this.viewer
      },

      isLocaleJa() {
        return window.AnnConfig.locale === 'ja'
      }
    },
  })
})

ujs.start()
Turbolinks.start()
