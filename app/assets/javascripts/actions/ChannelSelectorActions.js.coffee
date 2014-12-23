SelectorSpinnerActions = Annict.Actions.SelectorSpinnerActions

class Annict.Actions.ChannelSelectorActions
  constructor: (props) ->
    @workId = props.workId
    @currentChannelId = props.currentChannelId

  handleChange: (event) ->
    @newChannelId = event.target.value
    @selectedWorkId = $(event.target).data('workId')

    if @workId == @selectedWorkId && @_didChannelIdChange()
      SelectorSpinnerActions.show(@workId)

      $.ajax
        type: 'POST'
        url: "/api/works/#{@selectedWorkId}/channels/select"
        data:
          channel_id: @newChannelId
      .done =>
        @currentChannelId = @newChannelId
        SelectorSpinnerActions.hide(@workId)

  _didChannelIdChange: ->
    @currentChannelId != @newChannelId &&
    !(@currentChannelId == '' && @newChannelId == 'no_select') # 未選択状態で「ステータス」を選択してなかったら
