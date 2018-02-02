import _ from 'lodash'
import KeenTracking from 'keen-tracking'

const client = new KeenTracking({
  projectId: gon.keen.projectId,
  writeKey: gon.keen.writeKey,
})

export default {
  trackEvent(collectionName, action, data) {
    const basicData = {
      action: action,
      device: gon.user.device,
      page_category: gon.basic.pageCategory,
      request_uuid: gon.user.requestUUID,
      user_id: gon.user.userId,
    }

    client.recordEvent(collectionName, _.merge(basicData, data))
  },
}
