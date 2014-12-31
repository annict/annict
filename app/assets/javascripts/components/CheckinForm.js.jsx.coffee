If = Annict.Components.If

CheckinFormActions = Annict.Actions.CheckinFormActions
CheckinFormStore = Annict.Stores.CheckinFormStore

Annict.Components.CheckinForm = React.createClass
  getInitialState: ->
    CheckinFormStore.getState()

  componentDidMount: ->
    CheckinFormStore.addChangeListener(@_onChange)

  _onChange: ->
    @setState(CheckinFormStore.getState())

  expand: ->
    CheckinFormActions.expand()

  submitPath: ->
    "/works/#{@props.workId}/episodes/#{@props.episodeId}/checkins"

  submit: (event) ->
    event.preventDefault()
    CheckinFormActions.submit(@submitPath(), @refs)

  render: ->
    state = @state

    `<form className='checkin-form'>
      <div className='form-group'>
        <textarea
          className='form-control'
          ref='comment'
          onClick={this.expand}
          onTouchStart={this.expand}
          placeholder='見た感想を残してみませんか？'
          rows={state.textareaRows}>
        </textarea>
      </div>
      <div className='checkbox'>
        <label>
          <input ref='spoil' type='checkbox' value='1' />
          ネタバレを含む
        </label>
      </div>
      <If test={this.props.fromTwitter}>
        <div className='checkbox'>
          <label>
            <input ref='sharedTwitter' type='checkbox' value='1' />
            Twitterにシェアする
          </label>
        </div>
      </If>
      <div className='checkin-button text-center'>
        <button className='btn btn-primary' type='submit' onClick={this.submit}>
          <i className='fa fa-check'></i>
          チェックインする
        </button>
      </div>
    </form>`
