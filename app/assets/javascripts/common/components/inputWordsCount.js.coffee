Vue = require "vue/dist/vue"

module.exports =
  template: "#t-input-words-count"

  data: ->
    wordsCount: 0

  props:
    maxWordsCount:
      type: Number
      required: true

    inputName:
      type: String
      required: true

  mounted: ->
    $inputArea = $("form [name='#{@inputName}']")

    $inputArea.on "input", =>
      @wordsCount = $inputArea.val().length
