keen = require "../keen"

module.exports =
  template: "#t-favorite-button"

  props:
    resourceType:
      type: String
      required: true
    resourceId:
      type: Number
      required: true
    initIsFavorited:
      type: Boolean
      required: true
    isSignedIn:
      type: Boolean
      default: false

  data: ->
    isFavorited: @initIsFavorited
    isSaving: false

  computed:
    buttonText: ->
      if @isFavorited
        gon.I18n["messages._components.favorite_button.added_to_favorites"]
      else
        gon.I18n["messages._components.favorite_button.add_to_favorites"]

  methods:
    toggleFavorite: ->
      unless @isSignedIn
        $(".c-sign-up-modal").modal("show")
        keen.trackEvent("sign_up_modals", "open", via: "favorite_button")
        return

      @isSaving = true

      if @isFavorited
        $.ajax
          method: "POST"
          url: "/api/internal/favorites/unfavorite"
          data:
            resource_type: @resourceType
            resource_id: @resourceId
        .done =>
          @isFavorited = false
          @isSaving = false
      else
        $.ajax
          method: "POST"
          url: "/api/internal/favorites"
          data:
            resource_type: @resourceType
            resource_id: @resourceId
            page_category: gon.basic.pageCategory
        .done =>
          @isFavorited = true
          @isSaving = false
