module.exports =
  template: "#t-reaction-button"

  props:
    resourceType:
      type: String
      required: true
    resourceId:
      type: Number
      required: true
    initReactionsCount:
      type: Number
      required: true
    initIsReacted:
      type: Boolean
      required: true

  data: ->
    isSignedIn: gon.user.isSignedIn
    reactionsCount: @initReactionsCount
    isReacted: @initIsReacted
    isLoading: false

  methods:
    toggleReact: ->
      unless @isSignedIn
        $(".c-sign-up-modal").modal("show")
        return

      return if @isLoading

      @isLoading = true

      if @isReacted
        $.ajax
          method: "POST"
          url: "/api/internal/reactions/remove"
          data:
            resource_type: @resourceType
            resource_id: @resourceId
            kind: "thumbs_up"
        .done =>
          @reactionsCount += -1
          @isReacted = false
        .always =>
          @isLoading = false
      else
        $.ajax
          method: "POST"
          url: "/api/internal/reactions/add"
          data:
            resource_type: @resourceType
            resource_id: @resourceId
            kind: "thumbs_up"
            page_category: gon.basic.pageCategory
        .done =>
          @reactionsCount += 1
          @isReacted = true
        .always =>
          @isLoading = false
