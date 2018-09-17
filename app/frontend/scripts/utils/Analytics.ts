import { BaseData } from './BaseData'
import { ContextData } from './ContextData'

export default {
  load(baseData: BaseData, contextData: ContextData) {
    window.dataLayer = window.dataLayer || []
    function gtag(...args: any[]) {
      window.dataLayer.push(arguments)
    }
    const { ENCODED_USER_ID, USER_TYPE, VIEWER_UUID } = contextData
    const { GA_TRACKING_ID, PAGE_CATEGORY } = baseData

    gtag('js', new Date())
    gtag('config', GA_TRACKING_ID, {
      client_id: VIEWER_UUID,
      custom_map: {
        dimension1: USER_TYPE,
        dimension2: PAGE_CATEGORY,
      },
      dimension1: USER_TYPE,
      dimension2: PAGE_CATEGORY,
      user_id: ENCODED_USER_ID,
    })
  },
}
