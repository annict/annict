Annict.angular.config ($translateProvider) ->
  $translateProvider.translations 'ja',
    statuses:
      kinds:
        wanna_watch: '見たい'
        watching: '見てる'
        watched: '見た'
        stop_watching: '見るのやめた'
      no_select: 'ステータス'
    titles:
      timeline: 'タイムライン'
      profile: 'プロフィール'
      works: '作品を探す'
      works_search: '作品検索'
    users:
      follow: 'フォローする'
      following: 'フォロー中'

  $translateProvider.preferredLanguage('ja')
