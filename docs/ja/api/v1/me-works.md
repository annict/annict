# GET /v1/me/works

自分がステータスを設定している作品の情報を取得することができます。

## フィールド

[GET /v1/works](https://annict.wikihub.io/wiki/api/works#get-v1-works) のフィールドと同じです。


## パラメータ

| 名前 | 概要 | 使用例 |
| --- | --- | --- |
| fields | レスポンスボディに含まれるデータのフィールドを絞り込みます。 | fields=id,title |
| filter_ids | 作品を作品IDで絞り込みます。 | filter_ids=1,2,3 |
| filter_season | 作品をリリース時期で絞り込みます。`2016-all` としたときは、2016年にリリースされる作品全てを取得することができます。 | filter_season=2016-spring |
| filter_title | 作品をタイトルで絞り込みます。 | filter_title=shirobako |
| filter_status | 作品をステータスで絞り込みます。`wanna_watch, watching, watched, on_hold, stop_watching` が指定できます。 | filter_status=watching |
| page | ページ数を指定します。 | page=2 |
| per_page | 1ページに何件取得するかを指定します。デフォルトは `25` 件で、`50` 件まで指定できます。 | per_page=30 |
| sort_id | 作品を作品IDで並び替えます。`asc` または `desc` が指定できます。 | sort_id=desc |
| sort_season | 作品をリリース時期で並び替えます。`asc` または `desc` が指定できます。 | sort_season=desc |
| sort_watchers_count | 作品をWatchersの数で並び替えます。`asc` または `desc` が指定できます。 | sort_watchers_count=desc |


## リクエスト例

```
$ curl -X GET https://api.annict.com/v1/me/works?access_token=(access_token)
```

```json
{
  "works": [
    {
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
    }
  ],
...
```
