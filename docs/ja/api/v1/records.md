# GET /v1/records

エピソードへの記録を取得することができます。

## フィールド

| 名前 | 概要 |
| --- | --- |
| id | 記録ID |
| comment | 記録したときに書かれた感想 |
| rating | 記録したときに付けられたレーティング。`0` から `5` までの数値が入っています。 |
| is_modified | この記録が編集されたかどうか |
| likes_count | Likeされた数 |
| comments_count | コメントされた数 |
| created_at | 記録された日時 |
| user | この記録をした利用者の情報 |
| work | この記録が紐づく作品情報。取得できるフィールドは [Works](https://annict.wikihub.io/wiki/api/works) と同じです。 |
| episode | この記録が紐づくエピソード情報 |

## パラメータ

| 名前 | 概要 | 使用例 |
| --- | --- | --- |
| fields | レスポンスボディに含まれるデータのフィールドを絞り込みます。 | fields=id,title |
| filter_ids | 記録を記録IDで絞り込みます。 | filter_ids=1,2,3 |
| filter_episode_id | 記録をエピソードIDで絞り込みます。 | filter_episode_id=1111 |
| page | ページ数を指定します。 | page=2 |
| per_page | 1ページに何件取得するかを指定します。デフォルトは `25` 件で、`50` 件まで指定できます。 | per_page=30 |
| sort_id | 記録を記録IDで並び替えます。`asc` または `desc` が指定できます。 | sort_id=desc |
| sort_likes_count | 記録をLikeされた数で並び替えます。`asc` または `desc` が指定できます。 | sort_likes_count=desc |


## リクエスト例

```
$ curl -X GET https://api.annict.com/v1/records?filter_episode_id=74669&access_token=(access_token)
```

```json
{
  "records": [
    {
      "id": 425551,
      "comment": "ゆるふわ田舎アニメかと思ったらギャグと下ネタが多めのコメディアニメだった。これはこれで。日岡さんの声良いなあ。",
      "rating": 4,
      "is_modified": false,
      "likes_count": 0,
      "comments_count": 0,
      "created_at": "2016-04-11T14:19:13.974Z",
      "user": {
        "id": 2,
        "username": "shimbaco",
        "name": "Koji Shimba",
        "description": "アニメ好きが高じてこのサービスを作りました。聖地巡礼を年に数回しています。",
        "url": "http://shimba.co",
        "records_count": 1906,
        "created_at": "2014-03-02T15:38:40.000Z"
      },
      "work": {
        "id": 4670,
        "title": "くまみこ",
        "title_kana": "くまみこ",
        "media": "tv",
        "media_text": "TV",
        "season_name": "2016-spring",
        "season_name_text": "2016年春",
        "released_on": "",
        "released_on_about": "",
        "official_site_url": "http://kmmk.tv/",
        "wikipedia_url": "https://ja.wikipedia.org/wiki/%E3%81%8F%E3%81%BE%E3%81%BF%E3%81%93",
        "twitter_username": "kmmk_anime",
        "twitter_hashtag": "kumamiko",
        "episodes_count": 6,
        "watchers_count": 609
      },
      "episode": {
        "id": 74669,
        "number": "1",
        "number_text": "第壱話",
        "sort_number": 10,
        "title": "クマと少女 お別れの時",
        "records_count": 183
      }
    },
...
```
