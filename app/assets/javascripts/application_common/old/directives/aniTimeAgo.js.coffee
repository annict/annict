AnnictOld.angular.directive "aniTimeAgo", ->
  (scope, elm, attr) ->
    date = moment(attr.aniTimeAgo)
    passageDays = moment().diff(moment(date), "days")

    text = if passageDays > 3
      date.format("YYYY/MM/DD")
    else
      date.fromNow()

    elm.prop("title", date.format("YYYY/MM/DD HH:mm"))
    elm.text(text)
