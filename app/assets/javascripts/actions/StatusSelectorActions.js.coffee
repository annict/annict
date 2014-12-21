SelectorSpinnerActions = Annict.Actions.SelectorSpinnerActions

class Annict.Actions.StatusSelectorActions
  constructor: (props) ->
    @workId = props.workId
    @currentStatusKind = props.currentStatusKind

  handleChange: (event) ->
    @newStatusKind = event.target.value
    @selectedWorkId = $(event.target).data('workId')

    if @workId == @selectedWorkId && @_didStatusKindChange()
      SelectorSpinnerActions.show()

      $.ajax
        type: 'POST'
        url: "/works/#{@selectedWorkId}/statuses/select"
        data:
          status_kind: @newStatusKind
      .done (data) ->
        SelectorSpinnerActions.hide()

  _didStatusKindChange: ->
    @currentStatusKind != @newStatusKind &&
    !(@currentStatusKind == '' && @newStatusKind == 'no_select') # 未選択状態で「ステータス」を選択してなかったら
