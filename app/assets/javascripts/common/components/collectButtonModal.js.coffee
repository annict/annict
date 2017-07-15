eventHub = require "../../common/eventHub"

MODAL_MODE =
  LIST_COLLECTION: "listCollection"
  NEW_COLLECTION: "newCollection"

module.exports =
  template: "#t-collect-button-modal"

  data: ->
    collectionDescription: ""
    collections: []
    collectionTitle: ""
    isAddingToCollection: false
    isLoadingCollections: false
    isSavingCollection: false
    itemComment: ""
    itemTitle: ""
    mode: MODAL_MODE.LIST_COLLECTION
    workId: null
    workImageUrl: null

  methods:
    loadCollections: ->
      $.ajax
        method: "GET"
        url: "/api/internal/collections"
        data:
          work_id: @workId
      .done (data) =>
        @collections = data.collections
        @itemTitle = data.work.title
        @workImageUrl = data.work.image_url
      .fail ->
        message = gon.I18n["messages._components.collect_button.error"]
        eventHub.$emit "flash:show", message, "alert"
      .always =>
        @isLoadingCollections = false

    addToCollection: (collection) ->
      return if collection.is_contained

      @isAddingToCollection = true

      $.ajax
        method: "POST"
        url: "/api/internal/collections/#{collection.id}/collection_items"
        data:
          work_id: @workId
          title: @itemTitle
          comment: @itemComment
      .done (data) ->
        $(".c-collect-button-modal").modal("hide")
        collectionLink = """
          <a href='/@#{data.user.username}/collections/#{data.collection.id}'>
            #{gon.I18n["messages._components.collect_button_modal.view_collection"]}
          </a>
        """
        message = gon.I18n["messages._components.collect_button_modal.added"]
        eventHub.$emit "flash:show", "#{message} #{collectionLink}"
      .fail ->
        message = gon.I18n["messages._components.collect_button.error"]
        eventHub.$emit "flash:show", message, "alert"
      .always =>
        @isAddingToCollection = false

    newCollection: ->
      @mode = MODAL_MODE.NEW_COLLECTION

    createCollection: ->
      @isSavingCollection = true

      $.ajax
        method: "POST"
        url: "/api/internal/collections"
        data:
          title: @collectionTitle
          description: @collectionDescription
          work_id: @workId
      .done (data) =>
        @collections = data.collections
        @mode = MODAL_MODE.LIST_COLLECTION
      .fail ->
        message = gon.I18n["messages._components.collect_button.error"]
        eventHub.$emit "flash:show", message, "alert"
      .always =>
        @isSavingCollection = false

    cancelForm: ->
      @collectionTitle = @collectionDescription = ""
      @mode = MODAL_MODE.LIST_COLLECTION

  created: ->
    eventHub.$on "collectButtonModal:show", (workId) =>
      @workId = workId
      @isLoadingCollections = true
      $(".c-collect-button-modal").modal("show")
      @loadCollections()
