EpisodesActions = Annict.Actions.EpisodesActions
EpisodesStore = Annict.Stores.EpisodesStore

Annict.Components.Episodes = React.createClass
  getInitialState: ->
    EpisodesActions.setDefaultState(@props)
    EpisodesStore.getState()

  componentDidMount: ->
    EpisodesStore.addChangeListener(@_onChange)

  _onChange: ->
    @setState(EpisodesStore.getState())

  startMultipleCheckinMode: ->
    EpisodesActions.startMultipleCheckinMode()

  stopMultipleCheckinMode: ->
    EpisodesActions.stopMultipleCheckinMode()

  checkAll: ->
    EpisodesActions.checkAll(@getEpisodeIds())

  uncheckAll: ->
    EpisodesActions.uncheckAll()

  submit: ->
    EpisodesActions.submit(@props.workId, @state.checkedEpisodeIds)

  getEpisodeIds: ->
    _.pluck(@state.episodes, 'id')

  isCheckingAll: ->
    _.difference(@getEpisodeIds(), @state.checkedEpisodeIds).length == 0

  getEpisodes: ->
    props = @props
    state = @state
    checkedEpisodeIds = state.checkedEpisodeIds

    state.episodes.map (episode) ->
      episodePath = "/works/#{episode.workId}/episodes/#{episode.id}"

      `<Annict.Components.Episode
        key={episode.id}
        checkedEpisodeIds={checkedEpisodeIds}
        episode={episode}
        episodePath={episodePath}
        isSignedIn={props.isSignedIn}
        isMultipleCheckinMode={state.isMultipleCheckinMode}
      />`

  render: ->
    classSet = React.addons.classSet
    props = @props
    state = @state

    switchButtonClass = classSet
      'fake-link': true
      'fake-link-small': true
      hide: !props.isSignedIn || state.isMultipleCheckinMode
    checkAllButtonClass = classSet
      'fake-link': true
      'fake-link-small': true
      hidden: !props.isSignedIn || !state.isMultipleCheckinMode || @isCheckingAll()
    uncheckAllButtonClass = classSet
      'fake-link': true
      'fake-link-small': true
      hidden: !props.isSignedIn || !state.isMultipleCheckinMode || !@isCheckingAll()
    formClass = classSet
      hidden: !state.isMultipleCheckinMode
    checkIconClass = classSet
      fa: true
      hidden: @state.submitting
      'fa-check': true

    `<div className='work-episodes container'>
      <h2 className='text-center'>エピソード</h2>
      <div className='multiple-checkin'>
        <div className={switchButtonClass} onClick={this.startMultipleCheckinMode} onTouchStart={this.startMultipleCheckinMode}>一括チェックイン</div>
        <span className={checkAllButtonClass} onClick={this.checkAll}>全て選択</span>
        <span className={uncheckAllButtonClass} onClick={this.uncheckAll}>全て解除</span>
      </div>
      <table className='table table-striped'>
        <tbody>{this.getEpisodes()}</tbody>
      </table>
      <form className={formClass}>
        <div className='submit-menu'>
          <span className='fake-link fake-link-small' onClick={this.stopMultipleCheckinMode}>キャンセル</span>
          <button
            className='btn btn-submitting btn-primary btn-sm'
            disabled={(state.submitting || state.checkedEpisodeIds.length === 0) ? 'disabled' : false}
            onClick={this.submit}>
            <Annict.Components.Spinner target='multipleCheckin' />
            <i className={checkIconClass}></i>
            チェックインする
          </button>
        </div>
        <div className='clearfix'></div>
      </form>
      <div className='update-request'>
        最新のエピソードが登録されていない場合は、お手数ですが
        <a
          data-confirm='更新リクエストを送信しますか？'
          data-method='post'
          href={'/works/' + props.workId + '/appeals'}>
          更新リクエストを送信
        </a>
        してください。
        <br />
        更新の完了は<a href='https://twitter.com/anannict' target='_blank'>Twitter</a>でお知らせします。
      </div>
    </div>`
