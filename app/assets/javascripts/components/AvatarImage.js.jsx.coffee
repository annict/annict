Annict.Components.AvatarImage = React.createClass
  render: ->
    `<a className='profile-image pull-left' href={'/users/' + this.props.username}>
      <img className='img-circle' height={this.props.size} src={this.props.avatarUrl} width={this.props.size} />
     </a>`
