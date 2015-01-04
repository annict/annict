UserShareButtonActions = Annict.Actions.UserShareButtonActions
UserShareButtonStore = Annict.Stores.UserShareButtonStore

Annict.Components.UserShareButton = React.createClass
  getInitialState: ->
    UserShareButtonStore.getState()

  componentDidMount: ->
    UserShareButtonStore.addChangeListener(@_onChange)

  _onChange: ->
    @setState(UserShareButtonStore.getState())

  openModal: ->
    UserShareButtonActions.openModal(@props.potteUrl, @props.username)

  submit: ->
    UserShareButtonActions.submit(@)

  render: ->
    state = @state
    classSet = React.addons.classSet
    loadingClass = classSet(loading: true, hidden: @state.shareImageLoaded)
    thumbnailClass = classSet(hidden: !@state.shareImageLoaded)
    bodyCountClass = classSet('body-count': true, over: @state.isBodyCountOver)
    disabledSubmitButton = !@state.shareImageLoaded || @state.isBodyCountOver

    `<div className='share-button ani-share-button-modal' data-is-mobile="#{browser.mobile?}">
      <span className='btn btn-twitter' onClick={this.openModal}>
        <i className='fa fa-twitter'></i>
        見てるアニメをツイートする
      </span>
      <div className='modal fade' id='js-share-button-modal' tabindex='-1' role='dialog' aria-labelledby='modalLabel' aria-hidden='true'>
        <div className='modal-dialog'>
          <div className='modal-content'>
            <div className='modal-header'>
              <button className='close' type='button' data-dismiss='modal'>
                <span aria-hidden='true'>
                  <i className='fa fa-times'></i>
                </span>
              </button>
              <h4 className='modal-title'>ツイートする</h4>
            </div>
            <div className='modal-body'>
              <div className='description'>
                見てるアニメを画像付きでツイートします。実際にツイートされる画像は下のサンプル画像より大きいものになります。
              </div>
              <div className='image'>
                <div className={loadingClass}>
                  <Annict.Components.Spinner target='shareImageLoading' />
                  <div className='message'>画像を生成中...</div>
                </div>
                <img src={state.thumbnailUrl} className={thumbnailClass} width='200' height='200' />
              </div>
              <form>
                <textarea className='form-control' ref='body' onInput={UserShareButtonActions.countDownBody} placeholder='コメントを書いてツイートしよう！' rows='3'></textarea>
                <div className={bodyCountClass}>{state.bodyCount}</div>
                <button className='btn btn-twitter' onClick={this.submit} disabled={disabledSubmitButton ? 'disabled': false}>
                  <i className='fa fa-twitter'></i>
                  ツイートする
                </button>
              </form>
            </div>
          </div>
        </div>
      </div>
    </div>`
