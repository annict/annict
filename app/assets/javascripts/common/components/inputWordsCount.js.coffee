Vue = require "vue/dist/vue"

module.exports =
  template: "#t-input-words-count"

  data: ->
    wordsCount: @initWordsCount

  props:
    maxWordsCount:
      type: Number
      required: true

    initWordsCount:
      type: Number
      required: true

    inputName:
      type: String
      required: true

  mounted: ->
    $inputArea = $("form [name='#{@inputName}']")

    $inputArea.on "input", =>
      @wordsCount = $inputArea.val().length
