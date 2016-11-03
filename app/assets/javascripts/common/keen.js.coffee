_ = require "lodash"
KeenTracking = require "keen-tracking"

client = new KeenTracking
  projectId: gon.keen.projectId
  writeKey: gon.keen.writeKey

module.exports =
  trackEvent: (collectionName, action, data) ->
    basicData =
      action: action
      device: gon.user.device
      page_category: gon.basic.pageCategory
      client_uuid: gon.user.clientUUID
      user_id: gon.user.userId

    client.recordEvent(collectionName, _.merge(basicData, data))
