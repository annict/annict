ProgramsStore = Annict.Stores.ProgramsStore
ProgramsActions = Annict.Actions.ProgramsActions

Annict.Components.Programs = React.createClass
  getInitialState: ->
    ProgramsStore.getState()

  componentDidMount: ->
    ProgramsStore.addChangeListener(@_onChange)
    ProgramsActions.getPrograms()

  _onChange: ->
    @setState(ProgramsStore.getState())

  render: ->
    programs = @state.programs.map (program) ->
      `<Annict.Components.Program key={program.id} program={program} />`
    loader = `<Annict.Components.Loader loading={this.state.loading} />`

    if !@state.loading && _.isEmpty(programs)
      `<div className='info well'>
        <div className='icon'>
          <i className='fa fa-info-circle'></i>
        </div>
        <p>表示できる番組はありません。</p>
        <p>普段見ているテレビのチャンネルをAnnictに設定すると、このページにそのチャンネルで放送中のアニメが表示されます。</p>
        <p>チャンネルは<a href='/channels'>チャンネル一覧ページ</a>で設定できます。</p>
      </div>`
    else
      `<div className='programs'>
        <Annict.Components.InfiniteScroll loadMore={ProgramsActions.getPrograms} hasMore={this.state.hasMore} loader={loader}>
          {programs}
        </Annict.Components.InfiniteScroll>
      </div>`
