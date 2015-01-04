Annict.Components.ActivityWorkInfo = React.createClass
  render: ->
    activity = @props.activity

    `<div className='work-info'>
      <a href={'/works/' + activity.work.id}>
        <div className='work'>
          <div className='image'>
            <img alt={activity.work.title} height='40' src={activity.main_item.image_url} width='40' />
          </div>
          <div className='title'>{activity.work.title}</div>
          <div className='clearfix'></div>
        </div>
        <div className='clearfix'></div>
      </a>
    </div>`
