SpinnerStore = Annict.Stores.SpinnerStore

Annict.Components.Spinner = React.createClass
  getInitialState: ->
    SpinnerStore.getState()

  componentDidMount: ->
    SpinnerStore.addChangeListener(@_onChange)

  componentWillUnmount: ->
    SpinnerStore.removeChangeListener(@_onChange);

  _onChange: ->
    @setState(SpinnerStore.getState())

  render: ->
    classSet = React.addons.classSet

    spinnerClass = classSet
      spinner: true
      fa: true
      hidden: !_.contains(@state.visibleSpinners, @props.target)
      'fa-circle-o-notch': true
      'fa-spin': true

    `<i className={spinnerClass}></i>`
