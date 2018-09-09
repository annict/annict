import { BaseData } from './BaseData'
import { PageData } from './PageData'

export default {
  load(baseData: BaseData, pageData: PageData) {
    window.dataLayer = window.dataLayer || []
    function gtag(...args: any[]) {
      window.dataLayer.push(arguments)
    }
    const { encodedUserId, gaTrackingId, userType, viewerUUID } = baseData
    const { pageCategory } = pageData

    gtag('js', new Date())
    gtag('config', gaTrackingId, {
      client_id: viewerUUID,
      custom_map: {
        dimension1: userType,
        dimension2: pageCategory,
      },
      dimension1: userType,
      dimension2: pageCategory,
      user_id: encodedUserId,
    })
  },
}
