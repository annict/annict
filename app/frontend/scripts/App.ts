import axios from 'axios'
import * as Cookies from 'js-cookie'
import * as moment from 'moment-timezone'
import Vue from 'vue'
import Component from 'vue-class-component'

import AccessToken from './AccessToken'

@Component
export default class App extends Vue {
  private csrfParam = ''
  private csrfToken = ''
  private domain = ''
  private env = ''
  private isAppLoaded = false
  private isSignedIn = false
  private locale = ''

  get isProduction() {
    return this.env === 'production'
  }

  private async setup() {
    moment.locale(this.locale)

    Cookies.set('ann_time_zone', moment.tz.guess(), {
      domain: `.${this.domain}`,
      secure: this.isProduction,
    })

    axios.defaults.headers.common = {
      'X-CSRF-Token': this.csrfToken,
    }

    await AccessToken.generate()
  }

  private setupVue() {
    Vue.config.silent = this.isProduction
    Vue.config.devtools = !this.isProduction
    Vue.config.performance = !this.isProduction
    Vue.config.productionTip = !this.isProduction
  }

  private async created() {
    const baseData = await axios.get('/api/internal/v3/base_data')

    Object.assign(this, baseData)

    await this.setup()
    this.setupVue()

    this.isAppLoaded = true
  }
}
