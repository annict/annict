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
      notifications: {
        index: '通知',
      },
      pages: {
        about: 'Annictとは？',
        legal: '特定商取引法に基づく表記',
        privacy: 'プライバシーポリシー',
        terms: '利用規約'
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
        _notAdded: '登録されていません',
        noRecordBodyList: '感想はありません',
      },
      signUpModal: {
        body: 'ユーザ登録するとこの機能が使えます。'
      },
      statusSelector: {
        selectStatus: 'ステータスを選択',
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
    status: {
      kind: {
        planToWatch: '見たい',
        watching: '見てる',
        completed: '見た',
        onHold: '一時中断',
        dropped: '視聴中止',
      },
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
    completed: '見たアニメ',
    contents: 'コンテンツ',
    currentSeason: '今期のアニメ',
    delete: '削除',
    detail: '詳細',
    dropped: '視聴中止したアニメ',
    edit: '編集',
    english: '英語',
    episodes: 'エピソード',
    faqs: 'よくある質問',
    home: 'ホーム',
    information: '基本情報',
    japanese: '日本語',
    languages: '言語',
    library: 'ライブラリ',
    menu: 'メニュー',
    music: '音楽',
    myAnimeList: 'MyAnimeList',
    nextSeason: '来期のアニメ',
    onHold: '一時中断してるアニメ',
    overall: '全体',
    planToWatch: '見たいアニメ',
    prevSeason: '前期のアニメ',
    profile: 'プロフィール',
    slots: '放送予定',
    pv: 'PV',
    rating: '評価',
    ratingsCount: '評価数',
    recordBodyList: '感想',
    records: '記録',
    relatedWorks: '関連作品',
    releaseSeason: 'リリース時期',
    satisfactionRateShorten: '満足度',
    seasonalAnime: 'シーズン別アニメ',
    seasonXAnime: '{seasonName}アニメ',
    seriesWithName: '{seriesName}シリーズ',
    services: 'サービス',
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
    watching: '見てるアニメ',
    watchingShorten: '見てる',
    yearFall: '{year}年秋',
    yearSpring: '{year}年春',
    yearSummer: '{year}年夏',
    yearWinter: '{year}年冬'
  },
  verb: {
    explore: '見つける',
    search: '検索する',
    signOut: 'ログアウトする',
    track: '記録する',
  },
}
