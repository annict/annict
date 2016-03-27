Vue = require "vue"
moment = require "moment"

module.exports = Vue.extend
  template: """
    <span class='ann-time-ago' title='{{ timeAgoDetail }}'>
      {{ timeAgo }}
    </span>
  """

  props:
    time:
      type: String
      required: true

  data: ->
    datetime: moment(@time)

  computed:
    timeAgo: ->
      current = moment()
      date = @datetime.format("YYYY-MM-DD")
      currentDate = current.format("YYYY-MM-DD")

      passageDays = moment(currentDate).diff(moment(date), "days")

      if passageDays > 3
        @datetime.format("YYYY/MM/DD")
      else
        @datetime.fromNow()

    timeAgoDetail: ->
      @datetime.format("YYYY/MM/DD HH:mm")
