Annict.Components.ActivityCheckinComment = React.createClass
  getInitialState: ->
    spoilGuardClasses:
      'hide': !@props.checkin.spoil
    bodyClasses:
      'hide': @props.checkin.spoil
  hideSpoilGuard: ->
    @setState(spoilGuardClasses: { hide: true }, bodyClasses: { hide: false })
  render: ->
    if this.props.checkin.comment?.length > 0
      classSet = React.addons.classSet

      `<blockquote>
        <div className='checkin-comment'>
          <div className={'spoil-guard ' + classSet(this.state.spoilGuardClasses)} onClick={this.hideSpoilGuard}>
            <i className="fa fa-exclamation"></i>ネタバレを含んでいます (クリックで展開)
          </div>
          <div
            className={'body ' + classSet(this.state.bodyClasses)}
            dangerouslySetInnerHTML={{__html: Annict.Utils.simpleFormat(this.props.checkin.comment)}}
          />
        </div>
      </blockquote>`
    else
      false
