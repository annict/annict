eventHub = require "../../common/eventHub"

module.exports =
  template: "#t-amazon-item-attacher"

  props:
    resourceType:
      type: String
      required: true

    resourceId:
      type: String
      required: true

  data: ->
    keyword: ""
    page: 1
    items: []
    isLoading: false

  methods:
    search: (page) ->
      return unless @keyword

      $(window).scrollTop(0)

      @isLoading = true

      $.ajax
        method: "GET"
        url: "/api/internal/amazon/search"
        data:
          keyword: @keyword
          resource_type: @resourceType
          resource_id: @resourceId
          page: page
      .done (data) =>
        @items = data.items
        @page = page
        @totalPages = data.total_pages
      .fail ->
        message = gon.I18n["messages._components.amazon_item_attacher.error"]
        eventHub.$emit "flash:show", message, "alert"
      .always =>
        @isLoading = false

    add: (item) ->
      return if item.added_to_resource

      item.isLoading = true

      $.ajax
        method: "POST"
        url: "/api/internal/items"
        data:
          resource_type: @resourceType
          resource_id: @resourceId
          asin: item.asin
          page_category: gon.basic.pageCategory
      .done ->
        item.isLoading = false
        item.added_to_resource = true
