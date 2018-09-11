import axios from 'axios'
import 'bootstrap'
import * as Cookies from 'js-cookie'
import * as moment from 'moment-timezone'
import Vue from 'vue'
import Component from 'vue-class-component'

import AccessToken from './utils/AccessToken'
import Analytics from './utils/Analytics'
import './utils/Global'

import Home from './components/pages/Home'

@Component({
  components: {
    'c-home': Home,
  },
})
export default class App extends Vue {
  private csrfParam = ''
  private csrfToken = ''
  private domain = ''
  private encodedUserId = ''
  private env = ''
  private gaTrackingId = ''
  private isAppLoaded = false
  private isSignedIn = false
  private locale = ''
  private userType = ''
  private viewerUUID = ''

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
    const res = await axios.get('/api/internal/base_data')
    const baseData = res.data
    const pageData = window.ann.pageData

    Object.assign(this, baseData)

    Analytics.load(baseData, pageData)
    await this.setup()
    this.setupVue()

    this.isAppLoaded = true
  }
}
