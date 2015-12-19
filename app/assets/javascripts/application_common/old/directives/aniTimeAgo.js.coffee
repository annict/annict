AnnictOld.angular.directive "aniTimeAgo", ->
  (scope, elm, attr) ->
    datetime = moment(attr.aniTimeAgo)
    current = moment()
    date = datetime.format("YYYY-MM-DD")
    currentDate = current.format("YYYY-MM-DD")

    passageDays = moment(currentDate).diff(moment(date), "days")

    text = if passageDays > 3
      datetime.format("YYYY/MM/DD")
    else
      datetime.fromNow()

    elm.prop("title", datetime.format("YYYY/MM/DD HH:mm"))
    elm.text(text)
