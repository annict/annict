Annict.Components.Loader = React.createClass
  render: ->
    return false unless @props.loading

    `<div className='loading'><div className='core'>Loading...</div></div>`
