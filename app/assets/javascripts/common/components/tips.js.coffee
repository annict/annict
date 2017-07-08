module.exports =
  template: "#t-tips"

  data: ->
    tips: JSON.parse(gon.tips)

  methods:
    open: (index) ->
      tip = @tips[index]
      tip.isOpened = !tip.isOpened

    close: (index) ->
      if confirm gon.I18n["messages._common.are_you_sure"]
        $.ajax
          method: "POST"
          url: "/api/internal/tips/close"
          data:
            slug: @tips[index].slug
            page_category: gon.basic.pageCategory
        .done =>
          @tips.splice(index, 1)
