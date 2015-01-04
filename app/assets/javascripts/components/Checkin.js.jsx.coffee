If = Annict.Components.If

Annict.Components.Checkin = React.createClass
  render: ->
    props = @props

    checkinPath = "/works/#{props.workId}/episodes/#{props.episodeId}/checkins/#{props.checkin.id}"
    createdAt = Annict.Utils.timeAgo(props.checkin.created_at)

    `<div className='checkin'>
      <div className='media'>
        <div className='media-left'>
          <a className='pull-left' href='/users/{props.checkin.user.username}'>
            <img alt={props.checkin.profile.name} className='img-circle' height='50' src={props.checkin.profile.avatar_url} width='50' />
          </a>
        </div>
        <div className='media-body'>
          <div className='top'>
            <div className='pull-left'>
              <span className='name'>
                <a href='/users/{props.checkin.user.username}'>{props.checkin.profile.name}</a>
              </span>
            </div>
            <div className='pull-right'>
              <span className='created-at'>
                <a href={checkinPath}>{createdAt}</a>
              </span>
            </div>
          </div>
          <div className='middle'>
            <Annict.Components.CheckinComment comment={props.checkin.comment} spoil={props.checkin.spoil} />
          </div>
          <div className='bottom'>
            <div className='top'>
              <div className='pull-left'>
                <If test={props.checkin.twitter_click_count > 0}>
                  <span className='twitter-click-counter'>
                    <i className='fa fa-twitter'></i>{props.checkin.twitter_click_count}クリック
                  </span>
                </If>
              </div>
              <div className='pull-right'>
                <Annict.Components.LikeButton meta={props.checkin.meta} resource={props.checkin} resourceName='checkins' />
                <Annict.Components.CommentButton checkin={props.checkin} episodeId={props.episodeId} workId={props.workId} />
              </div>
            </div>
            <If test={props.currentUser.username === props.checkin.user.username}>
              <div className='bottom'>
                <div className='pull-right'>
                  <a className='edit-button' href={checkinPath + '/edit'}>
                    <i className='fa fa-edit'></i>編集
                  </a>
                  <a className='delete-button' data-confirm='チェックインを削除します。よろしいですか？' data-method='delete' href={checkinPath} rel='nofollow'>
                    <i className="fa fa-trash-o"></i>削除
                  </a>
                </div>
              </div>
            </If>
          </div>
        </div>
      </div>
      <hr />
    </div>`
