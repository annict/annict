export default {
  head: {
    title: {
      channels: {
        index: 'チャンネル一覧',
      },
      faqs: {
        index: 'よくある質問',
      },
      friends: {
        index: '友達を探す',
      },
      pages: {
        about: 'Annictについて',
      },
      workRecords: {
        show: 'アニメ「{workTitle}」の記録 by {profileName} ({username})'
      },
      works: {
        newest: '新規登録アニメ',
        popular: '人気アニメ',
      },
    },
  },
  messages: {
    _common: {
      searchWithKeywords: '作品名や人物名などで検索...',
    },
    _components: {
      empty: {
        noRecordBodyList: '感想はありません'
      }
    },
    works: {
      viewAllNRecordBodyList: '%{n}件の感想を全て見る'
    },
  },
  models: {
    record: {
      ratingState: {
        great: 'とても良い',
        good: '良い',
        average: '普通',
        bad: '良くない',
      }
    },
    season: {
      yearly: {
        all: '{year}年',
        winter: '{year}年冬',
        spring: '{year}年春',
        summer: '{year}年夏',
        autumn: '{year}年秋',
      },
      later: '時期未定',
    },
    work: {
      media: 'メディア',
      officialSiteUrl: '公式サイト',
      officialSiteUrlEn: '公式サイト (英語)',
      synopsis: 'あらすじ',
      titleEn: 'タイトル (英語)',
      titleKana: 'タイトル (かな)',
      twitterHashtag: 'ハッシュタグ',
      twitterUsername: '公式Twitter',
      wikipediaUrl: 'Wikipedia',
      wikipediaUrlEn: 'Wikipedia (英語)',
    },
  },
  noun: {
    about: 'About',
    airing: '放送中',
    animation: '映像',
    annictDb: 'Annict DB',
    annictDevelopers: 'Annict Developers',
    annictForum: 'Annict Forum',
    annictSupporters: 'Annict Supporters',
    annictUserland: 'Annict Userland',
    character: 'キャラクター',
    characters: 'キャラクター',
    currentSeason: '今期のアニメ',
    delete: '削除',
    detail: '詳細',
    edit: '編集',
    episodes: 'エピソード',
    home: 'ホーム',
    information: '基本情報',
    menu: 'メニュー',
    music: '音楽',
    myAnimeList: 'MyAnimeList',
    nextSeason: '来期のアニメ',
    overall: '全体',
    prevSeason: '前期のアニメ',
    profile: 'プロフィール',
    programs: '放送予定',
    pv: 'PV',
    rating: '評価',
    ratingsCount: '評価数',
    recordBodyList: '感想',
    records: '記録',
    releaseSeason: 'リリース時期',
    satisfactionRateShorten: '満足度',
    settings: '設定',
    share: 'シェア',
    signIn: 'ログイン',
    signUp: 'ユーザ登録',
    signUpShorten: '登録',
    source: '引用元',
    staffs: 'スタッフ',
    startToBroadcastMovieDate: '公開日',
    startToBroadcastTvDate: '放送開始日',
    startToPublishDate: '公開日',
    startToSellDate: '発売開始日',
    stats: '統計',
    story: 'ストーリー',
    supporter: 'サポーター',
    syoboiCalendar: 'しょぼいカレンダー',
    tweet: 'ツイート',
    vods: '動画サービス',
    watchersCount: '視聴者数',
    watchingShorten: '見てる',
  },
  verb: {
    explore: '見つける',
    search: '検索する',
    signOut: 'ログアウトする',
    track: '記録する',
  },
}
