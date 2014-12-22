SelectorSpinnerStore = Annict.Stores.SelectorSpinnerStore
SelectorSpinnerConstants = Annict.Constants.SelectorSpinnerConstants
SelectorSpinnerActions = Annict.Actions.SelectorSpinnerActions

Annict.Components.SelectorSpinner = React.createClass
  getInitialState: ->
    SelectorSpinnerStore.getState()

  componentDidMount: ->
    SelectorSpinnerStore.addChangeListener(@_onChange)

  componentWillUnmount: ->
    SelectorSpinnerStore.removeChangeListener(@_onChange);

  _onChange: ->
    state = SelectorSpinnerStore.getState()

    $spinner = $(@getDOMNode())
    spinnerColor = if @props.isMini then '#000000' else '#ffffff'
    spinnerOptions = { color: spinnerColor, lines: 8, length: 3, radius: 3, width: 1 }

    if state.hidden
      $circle = $spinner.find('i')
      $spinner.spin(false)
      $circle.removeClass('hidden')

      setTimeout ->
        $circle.addClass('hidden')
      , 2000
    else
      $spinner.spin(spinnerOptions)

  render: ->
    classSet = React.addons.classSet

    `<span className='selector-spinner' data-work-id={this.props.workId}>
      <i className='fa fa-check-circle hidden'></i>
    </span>`
