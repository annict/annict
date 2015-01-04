Annict.Components.InfiniteScroll = React.createClass
  displayName: 'InfiniteScroll'
  propTypes:
    pageStart: React.PropTypes.number
    threshold: React.PropTypes.number
    loadMore: React.PropTypes.func.isRequired
    hasMore: React.PropTypes.bool

  getDefaultProps: ->
    pageStart: 1
    hasMore: false
    threshold: 250

  componentDidMount: ->
    @pageLoaded = this.props.pageStart
    @attachScrollListener()

  componentDidUpdate: ->
    @attachScrollListener()

  render: ->
    props = @props

    `<div>{props.children}{props.loader}</div>`

  scrollListener: ->
    el = @getDOMNode()
    scrollTop = $(window).scrollTop()
    top = $(el).offset().top
    height = $(el).height()
    innerHeight = $(window).innerHeight()

    if (top + height - scrollTop - innerHeight) < Number(@props.threshold)
      @detachScrollListener()

      _.bind(@props.loadMore, @getLoadMoreContext(), @pageLoaded += 1)()

  getLoadMoreContext: ->
    if _.isEmpty(@props.loadMoreContext)
      @props
    else
      @props.loadMoreContext

  attachScrollListener: ->
    return if !@props.hasMore
    $(window).on('scroll', @scrollListener)
    $(window).on('resize', @scrollListener)
    @scrollListener()

  detachScrollListener: ->
    $(window).off('scroll', @scrollListener)
    $(window).off('resize', @scrollListener)

  componentWillUnmount: ->
    @detachScrollListener()
