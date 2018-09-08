import 'select2'
import 'd3'
import 'dropzone'
import Cookies from 'js-cookie'
import moment from 'moment-timezone'
import Turbolinks from 'turbolinks'
import Vue from 'vue'
import VueLazyload from 'vue-lazyload'

import AccessToken from './AccessToken'
import Ajax from './Ajax'

import Home from './components/Home'

document.addEventListener('turbolinks:load', async function() {
  Vue.use(VueLazyload)

  Vue.component('ann-home', Home)

  new Vue({
    el: '.p-application',

    data() {
      return {
        csrfParam: '',
        csrfToken: '',
        domain: '',
        env: '',
        locale: '',
        isSignedIn: false,
        isAppLoaded: false
      }
    },

    computed: {
      isProduction() {
        return this.env === 'production'
      }
    },

    methods: {
      async setup() {
        moment.locale(this.locale)

        Cookies.set('ann_time_zone', moment.tz.guess(), {
          domain: `.${this.domain}`,
          secure: this.isProduction,
        })

        Ajax.setup({
          headers: { 'X-CSRF-Token': this.csrfToken },
        })

        await AccessToken.generate()
      },

      setupVue() {
        Vue.config.silent = this.isProduction
        Vue.config.devtools = !this.isProduction
        Vue.config.performance = !this.isProduction
        Vue.config.productionTip = !this.isProduction
      },
    },

    async created() {
      const res = await fetch('/api/internal/v3/base_data')
      const baseData = await res.json()

      Object.assign(this, baseData)

      await this.setup()
      this.setupVue()

      this.isAppLoaded = true
    },
  })
})

Turbolinks.start()
