SelectorSpinnerStore = Annict.Stores.SelectorSpinnerStore
SelectorSpinnerActions = Annict.Actions.SelectorSpinnerActions

Annict.Components.SelectorSpinner = React.createClass
  getInitialState: ->
    SelectorSpinnerStore.getState()

  componentDidMount: ->
    SelectorSpinnerStore.addChangeListener(@_onChange)

  componentWillUnmount: ->
    SelectorSpinnerStore.removeChangeListener(@_onChange)

  _onChange: ->
    @setState(SelectorSpinnerStore.getState())

  render: ->
    classSet = React.addons.classSet

    spinnerClass = classSet
      fa: true
      spinner: true
      hidden: !_.contains(@state.spinningTargets, @props.target)
      'fa-circle-o-notch': true
      'fa-spin': true
    checkClass = classSet
      fa: true
      check: true
      hidden: !_.contains(@state.doneTargets, @props.target)
      'fa-check-circle': true

    `<span className='selector-spinner'>
      <i className={spinnerClass}></i>
      <i className={checkClass}></i>
    </span>`
