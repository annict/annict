_ = require "lodash"

eventHub = require "../../common/eventHub"

module.exports =
  template: "#t-impression-button-modal"

  data: ->
    workId: null
    tagNames: []
    allTagNames: []
    draftTagNames: []
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
        @tagNames = @draftTagNames = data.tag_names
        @allTagNames = data.all_tag_names
        @comment = data.comment

        setTimeout =>
          $tagsInput = $(".js-impression-tags")
          $tagsInput.select2
            tags: true
          $tagsInput.on "select2:select", (event) =>
            @draftTagNames = $(event.currentTarget).val()
          $tagsInput.on "select2:unselect", (event) =>
            @draftTagNames = $(event.currentTarget).val()
      .fail ->
        message = gon.I18n["messages._components.impression_button.error"]
        eventHub.$emit "flash:show", message, "alert"
      .always =>
        @isLoading = false

    save: ->
      @isSaving = true
      @tagNames = @draftTagNames

      $.ajax
        method: "PATCH"
        url: "/api/internal/impression"
        data:
          work_id: @workId
          tags: @draftTagNames
          comment: @comment
      .done (data) =>
        $(".c-impression-button-modal").modal("hide")
        eventHub.$emit "workTags:saved", @workId, data.tags
        eventHub.$emit "workComment:saved", @workId, data.comment
        eventHub.$emit "flash:show", gon.I18n["messages._common.updated"]
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
