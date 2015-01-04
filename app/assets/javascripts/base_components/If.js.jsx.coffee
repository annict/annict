Annict.Components.If = React.createClass
  render: ->
    if @props.test
      @props.children
    else
      false
