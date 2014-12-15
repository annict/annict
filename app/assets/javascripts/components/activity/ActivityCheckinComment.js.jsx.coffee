Annict.Components.ActivityCheckinComment = React.createClass
  getInitialState: ->
    spoilGuardClasses:
      'hide': !@props.activity.checkin.spoil
    bodyClasses:
      'hide': @props.activity.checkin.spoil
  hideSpoilGuard: ->
    @setState(spoilGuardClasses: { hide: true }, bodyClasses: { hide: false })
  render: ->
    activity = @props.activity

    if activity.checkin.comment?.length > 0
      classSet = React.addons.classSet

      `<blockquote>
        <div className='checkin-comment'>
          <div className={'spoil-guard ' + classSet(this.state.spoilGuardClasses)} onClick={this.hideSpoilGuard}>
            <i className="fa fa-exclamation"></i>ネタバレを含んでいます (クリックで展開)
          </div>
          <div
            className={'body ' + classSet(this.state.bodyClasses)}
            dangerouslySetInnerHTML={{__html: Annict.Utils.simpleFormat(activity.checkin.comment)}}
          />
        </div>
      </blockquote>`
    else
      false
