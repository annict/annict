Annict.Components.Program = React.createClass
  render: ->
    program = @props.program
    startedAt = moment(program.started_at).format('MM/DD HH:mm')

    `<div className='program'>
      <div className='media'>
        <a className='media-left' href={'/works/' + program.work.id}>
          <img alt={program.work.title} className='media-object' height='70' src={program.work.image_url} width='70' />
        </a>
        <div className='media-body'>
          <div className='work-title'>
            <a href={'/works/' + program.work.id}>{program.work.title}</a>
          </div>
          <div className='episode-title'>
            <a href={'/works/' + program.work.id + '/episodes/' + program.episode.id}>{program.episode.number}「{program.episode.title}」</a>
          </div>
          <div className='footer'>
            <div className='channel-name'>{program.channel.name}</div>
            <div className='started-at'>{startedAt} ~</div>
            <div className='clearfix'></div>
          </div>
        </div>
      </div>
      <hr />
    </div>`
