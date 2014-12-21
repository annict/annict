SelectorSpinnerStore = Annict.Stores.SelectorSpinnerStore
SelectorSpinnerConstants = Annict.Constants.SelectorSpinnerConstants
SelectorSpinnerActions = Annict.Actions.SelectorSpinnerActions

Annict.Components.SelectorSpinner = React.createClass
  getInitialState: ->
    SelectorSpinnerStore.getState()

  componentDidMount: ->
    SelectorSpinnerStore.addChangeListener(@_onChange)

  _onChange: ->
    @setState(SelectorSpinnerStore.getState())

    $spinner = $(@getDOMNode())
    spinnerColor = if @props.isMini then '#000000' else '#ffffff'
    spinnerOptions = { color: spinnerColor, lines: 8, length: 3, radius: 3, width: 1, zIndex: 1 }

    if @state.hidden
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

    spinnerClasses = classSet(spinner: true, hidden: @state.hidden)

    `<span className='selector-spinner' data-work-id={this.props.workId}>
      <span className={spinnerClasses}></span>
      <i className='fa fa-check-circle hidden'></i>
    </span>`
