Vue = require "vue/dist/vue"

keen = require "../keen"

module.exports = Vue.extend
  template: "#t-tips"

  data: ->
    tips: JSON.parse(gon.tips)

  methods:
    open: (index) ->
      tip = @tips[index]
      keen.trackEvent("tips", "open", slug: tip.slug) unless tip.isOpened
      tip.isOpened = !tip.isOpened

    close: (index) ->
      if confirm gon.I18n["messages._common.are_you_sure"]
        $.ajax
          method: "POST"
          url: "/api/internal/tips/close"
          data:
            slug: @tips[index].slug
        .done =>
          @tips.splice(index, 1)
