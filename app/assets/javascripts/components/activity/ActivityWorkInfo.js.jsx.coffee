Annict.Components.ActivityWorkInfo = React.createClass
  render: ->
    `<div className='work-info'>
      <a href={'/works/' + this.props.work.id}>
        <div className='work'>
          <div className='image'>
            <img alt={this.props.work.title} height='40' src={this.props.item.image_url} width='40' />
          </div>
          <div className='title'>{this.props.work.title}</div>
          <div className='clearfix'></div>
        </div>
        <div className='clearfix'></div>
      </a>
    </div>`
