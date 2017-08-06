eventHub = require "../../common/eventHub"

module.exports =
  template: "#t-impression-button-modal"

  data: ->
    workId: null
    tags: []
    allTags: []
    draftTags: []
    comment: ""
    isLoading: false
    isSaving: false

  methods:
    load: ->
      @isLoading = true

      $.ajax
        method: "GET"
        url: "/api/internal/impression"
        data:
          work_id: @workId
      .done (data) =>
        {@tags, @comment} = data
        @allTags = data.all_tags
        setTimeout =>
          $tagsInput = $(".js-impression-tags")
          $tagsInput.select2
            tags: true
          $tagsInput.on "select2:select", (event) =>
            @draftTags = $(event.currentTarget).val()
          $tagsInput.on "select2:unselect", (event) =>
            @draftTags = $(event.currentTarget).val()
      .fail ->
        message = gon.I18n["messages._components.impression_button.error"]
        eventHub.$emit "flash:show", message, "alert"
      .always =>
        @isLoading = false

    save: ->
      @isSaving = true

      $.ajax
        method: "PATCH"
        url: "/api/internal/impression"
        data:
          work_id: @workId
          tags: @draftTags
          comment: @comment
      .done (data) ->
        location.reload()
      .fail ->
        message = gon.I18n["messages._components.impression_button.error"]
        eventHub.$emit "flash:show", message, "alert"
      .always =>
        @isSaving = false

  created: ->
    eventHub.$on "impressionButtonModal:show", (workId) =>
      @workId = workId
      @load()
      $(".c-impression-button-modal").modal("show")
