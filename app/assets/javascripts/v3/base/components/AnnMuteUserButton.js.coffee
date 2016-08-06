Vue = require "vue"

module.exports = Vue.extend
  template: "#ann-mute-user-button"

  props:
    userId:
      type: String
      required: true

  methods:
    mute: ->
      if confirm("ミュートするとその人の記録がエピソードページやアクティビティから見えなくなります。\nミュートしますか？")
        $.ajax
          method: "POST"
          url: "/api/internal/mute_users"
          data:
            user_id: @userId
        .done =>
          @$dispatch "AnnMuteUser:mute", @userId
          console.log "done"
