SpinnerStore = Annict.Stores.SpinnerStore

Annict.Components.Spinner = React.createClass
  componentDidMount: ->
    SpinnerStore.addChangeListener(@_onChange)

  componentWillUnmount: ->
    SpinnerStore.removeChangeListener(@_onChange);

  _onChange: ->
    state = SpinnerStore.getState()

    $spinner = $(@getDOMNode())
    spinnerOptions = { color: '#000000', lines: 8, length: 3, radius: 3, width: 1 }

    if state.hidden
      $spinner.spin(false)
    else
      $spinner.spin(spinnerOptions)

  render: ->
    `<span className='spinner' data-target={this.props.target}></span>`
