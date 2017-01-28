# GET /v1/me/programs

放送予定を取得することができます。

## フィールド

| 名前 | 概要 |
| --- | --- |
| id | 放送予定ID |
| started_at | 放送開始日時 |
| is_rebroadcast | この放送予定が再放送かどうか。再放送の場合は `true` が、そうでない場合は `false` が格納されています。 |
| channel | チャンネル情報 |
| work | この放送予定が紐づく作品情報。取得できるフィールドは [Works](https://annict.wikihub.io/wiki/api/works) と同じです。 |
| episode | この放送予定が紐づくエピソード情報。取得できるフィールドは [Episodes](https://annict.wikihub.io/wiki/api/episodes) と同じです。 |


## パラメータ

| 名前 | 概要 | 使用例 |
| --- | --- | --- |
| fields | レスポンスボディに含まれるデータのフィールドを絞り込みます。 | fields=id,title |
| filter_ids | 放送予定を放送予定IDで絞り込みます。 | filter_ids=1,2,3 |
| filter_channel_ids | 放送予定をチャンネルIDで絞り込みます。 | filter_channel_ids=1,2,3 |
| filter_work_ids | 放送予定を作品IDで絞り込みます。 | filter_work_ids=1,2,3 |
| filter_started_at_gt | 放送予定を放送開始日時で絞り込みます。指定した日時以降の放送予定が取得できます。 | filter_started_at_gt=2016/05/06 21:10 |
| filter_started_at_lt | 放送予定を放送開始日時で絞り込みます。指定した日時以前の放送予定が取得できます。 | filter_started_at_lt=2016/05/06 21:10 |
| filter_unwatched | 未視聴の放送予定だけを取得します。 | filter_unwatched=true |
| filter_rebroadcast | 放送予定を再放送フラグをもとに絞り込みます。`true` を渡すと再放送だけが、`false` を渡すと再放送以外の放送予定が取得できます。 | filter_rebroadcast=true |
| page | ページ数を指定します。 | page=2 |
| per_page | 1ページに何件取得するかを指定します。デフォルトは `25` 件で、`50` 件まで指定できます。 | per_page=30 |
| sort_id | 放送予定を放送予定IDで並び替えます。`asc` または `desc` が指定できます。 | sort_id=desc |
| sort_started_at | 放送予定を放送開始日時で並び替えます。`asc` または `desc` が指定できます。 | sort_started_at=desc |


## リクエスト例

```
$ curl -X GET https://api.annict.com/v1/me/programs?sort_started_at=desc&filter_started_at_gt=2016/05/05 02:00&access_token=(access_token)
```

```json
{
  "programs": [
    {
      "id": 35387,
      "started_at": "2016-05-07T20:10:00.000Z",
      "is_rebroadcast": false,
      "channel": {
        "id": 4,
        "name": "日本テレビ"
      },
      "work": {
        "id": 4681,
        "title": "ふらいんぐうぃっち",
        "title_kana": "ふらいんぐうぃっち",
        "media": "tv",
        "media_text": "TV",
        "season_name": "2016-spring",
        "season_name_text": "2016年春",
        "released_on": "",
        "released_on_about": "",
        "official_site_url": "http://www.flyingwitch.jp/",
        "wikipedia_url": "https://ja.wikipedia.org/wiki/%E3%81%B5%E3%82%89%E3%81%84%E3%82%93%E3%81%90%E3%81%86%E3%81%83%E3%81%A3%E3%81%A1",
        "twitter_username": "flying_tv",
        "twitter_hashtag": "flyingwitch",
        "episodes_count": 5,
        "watchers_count": 695
      },
      "episode": {
        "id": 75187,
        "number": "5",
        "number_text": "第5話",
        "sort_number": 50,
        "title": "使い魔の活用法",
        "records_count": 0
      }
    },
...
```
