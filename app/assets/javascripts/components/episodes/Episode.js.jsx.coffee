If = Annict.Components.If

EpisodesActions = Annict.Actions.EpisodesActions

Annict.Components.Episode = React.createClass
  toggle: ->
    episodeId = $(@getDOMNode()).find('.checkbox-column').data('episode-id')
    EpisodesActions.toggle(episodeId, @isChecked())

  isChecked: ->
    _.contains(@props.checkedEpisodeIds, @props.episode.id)

  render: ->
    classSet = React.addons.classSet
    props = @props
    episode = props.episode

    episodeClass = classSet
      'checkbox-column': true
      hidden: !props.isMultipleCheckinMode

    `<tr>
      <If test={props.isSignedIn}>
        <td className={episodeClass} onClick={this.toggle} data-episode-id={episode.id}>
          <input type='checkbox' checked={this.isChecked()} readOnly='true' />
        </td>
      </If>
      <If test={episode.number && episode.title}>
        <td className='number'>
          <a href={props.episodePath}>{episode.number}</a>
        </td>
      </If>
      <If test={episode.number && episode.title}>
        <td className='title'>
          <a href={props.episodePath}>{episode.title}</a>
        </td>
      </If>
      <If test={!episode.number}>
        <td className='title'>
          <a href={props.episodePath}>{episode.workTitle}</a>
        </td>
      </If>
      <If test={props.isSignedIn}>
        <td className='checkins'>
          <If test={episode.userCheckinsCount >= 1}>
            <i className='fa fa-check'></i>
          </If>
          <If test={episode.userCheckinsCount > 1}>
            <span className='badge'>{episode.userCheckinsCount}</span>
          </If>
        </td>
      </If>
    </tr>`
