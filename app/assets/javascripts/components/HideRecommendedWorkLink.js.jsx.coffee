HideRecommendedWorkLinkActions = Annict.Actions.HideRecommendedWorkLinkActions

Annict.Components.HideRecommendedWorkLink = React.createClass
  hide: ->
    HideRecommendedWorkLinkActions.hide(@props.workId)

  render: ->
    `<div className='hide-work fake-link' onClick={this.hide}>
      表示しない
    </div>`
