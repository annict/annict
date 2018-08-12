import Cookies from 'js-cookie'
import moment from 'moment-timezone'
import 'moment/locale/ja'

import AccessToken from './AccessToken'
import Ajax from './Ajax'
import BaseData from './BaseData'

export default {
  async setup() {
    const csrfToken = await BaseData.fetch('csrfToken')
    const domain = await BaseData.fetch('domain')
    const env = await BaseData.fetch('env')
    const locale = await BaseData.fetch('locale')

    moment.locale(locale)
    Cookies.set('ann_time_zone', moment.tz.guess(), {
      domain: `.${domain}`,
      secure: env === 'production',
    })

    Ajax.setup({
      headers: { 'X-CSRF-Token': csrfToken },
    })

    await AccessToken.generate()
  },

  async isProduction() {
    return (await BaseData.fetch('env')) === 'production'
  },
}
