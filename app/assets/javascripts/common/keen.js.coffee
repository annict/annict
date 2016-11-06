_ = require "lodash"
KeenTracking = require "keen-tracking"

client = new KeenTracking
  projectId: gon.keen.projectId
  writeKey: gon.keen.writeKey

module.exports =
  trackEvent: (collectionName, action, data) ->
    basicData =
      action: action
      client_uuid: gon.user.clientUUID
      device: gon.user.device
      locale: gon.user.locale
      page_category: gon.basic.pageCategory
      user_id: gon.user.userId

    client.recordEvent(collectionName, _.merge(basicData, data))
