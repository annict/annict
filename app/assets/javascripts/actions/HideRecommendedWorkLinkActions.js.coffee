HideRecommendedWorkLinkConstants = Annict.Constants.HideRecommendedWorkLinkConstants

Annict.Actions.HideRecommendedWorkLinkActions =
  hide: (workId) ->
    $.ajax
      type: 'POST'
      url: "/api/works/#{workId}/hide"
    .done ->
      # やってはいけないことをしたと思っている。反省している。
      $('.works-list').find("[data-work-id=#{workId}]").hide()
