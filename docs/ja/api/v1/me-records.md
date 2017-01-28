# POST /v1/me/records

エピソードへの記録が作成できます。

**このリクエストには `write` スコープが必要になります。**

## パラメータ

| 名前 | 概要 | データ例 |
| --- | --- | --- |
| episode_id | **[必須]** エピソードID | 1234 |
| comment | 感想 | あぁ^～心がぴょんぴょんするんじゃぁ^～ |
| rating | レーティング | 4.5 |
| share_twitter | 記録をTwitterにシェアするかどうか。`true` または `false` が入力できます。指定しなかったときは `false` (シェアしない) になります。 | true |
| share_facebook | 記録をFacebookにシェアするかどうか。`true` または `false` が入力できます。指定しなかったときは `false` (シェアしない) になります。 | true |


## リクエスト例

```
$ curl -X POST https://api.annict.com/v1/me/records?episode_id=5013&comment=あぁ^～心がぴょんぴょんするんじゃぁ^～&access_token=(access_token)
```

```json
{
  "id": 470491,
  "comment": "あぁ^～心がぴょんぴょんするんじゃぁ^～",
  "rating": null,
  "is_modified": false,
  "likes_count": 0,
  "comments_count": 0,
  "created_at": "2016-05-07T09:40:32.159Z",
  "user": {
    "id": 2,
    "username": "shimbaco",
    "name": "Koji Shimba",
    "description": "",
    "url": null,
    "records_count": 123,
    "created_at": "2016-05-03T19:06:59.929Z"
  },
  "work": {
    "id": 3994,
    "title": "ご注文はうさぎですか？",
    "title_kana": "ごちゅうもんはうさぎですか",
    "media": "tv",
    "media_text": "TV",
    "season_name": "2014-spring",
    "season_name_text": "2014年春",
    "released_on": "2014-04-10",
    "released_on_about": "",
    "official_site_url": "http://www.gochiusa.com/",
    "wikipedia_url": "http://ja.wikipedia.org/wiki/%E3%81%94%E6%B3%A8%E6%96%87%E3%81%AF%E3%81%86%E3%81%95%E3%81%8E%E3%81%A7%E3%81%99%E3%81%8B%3F#.E3.83.86.E3.83.AC.E3.83.93.E3.82.A2.E3.83.8B.E3.83.A1",
    "twitter_username": "usagi_anime",
    "twitter_hashtag": "gochiusa",
    "episodes_count": 12,
    "watchers_count": 850
  },
  "episode": {
    "id": 5013,
    "number": null,
    "number_text": "第1羽",
    "sort_number": 1,
    "title": "ひと目で尋常でないもふもふだと見抜いたよ",
    "records_count": 103
  }
}
```


# PATCH /v1/me/records/:id

作成した記録を編集することができます。

**このリクエストには `write` スコープが必要になります。**

## パラメータ

| 名前 | 概要 | データ例 |
| --- | --- | --- |
| id | **[必須]** 記録ID | 1016 |
| comment | 感想 | あぁ^～心がぴょんぴょんするんじゃぁ^～ |
| rating | レーティング | 5.0 |
| share_twitter | 記録をTwitterにシェアするかどうか。`true` または `false` が入力できます。指定しなかったときは `false` (シェアしない) になります。 | true |
| share_facebook | 記録をFacebookにシェアするかどうか。`true` または `false` が入力できます。指定しなかったときは `false` (シェアしない) になります。 | true |


## リクエスト例

```
$ curl -X PATCH https://api.annict.com/v1/me/records/1016?comment=あぁ^～心がぴょんぴょんするんじゃぁ^～&rating=5.0&share_facebook=true&access_token=(access_token)
```

```json
{
  "id": 1016,
  "comment": "あぁ^～心がぴょんぴょんするんじゃぁ^～",
  "rating": 5.0,
  "is_modified": true,
  "likes_count": 0,
  "comments_count": 0,
  "created_at": "2016-05-07T09:40:32.159Z",
  "user": {
    "id": 2,
    "username": "shimbaco",
    "name": "Koji Shimba",
    "description": "",
    "url": null,
    "records_count": 1234,
    "created_at": "2016-05-03T19:06:59.929Z"
  },
  "work": {
    "id": 3994,
    "title": "ご注文はうさぎですか？",
    "title_kana": "ごちゅうもんはうさぎですか",
    "media": "tv",
    "media_text": "TV",
    "season_name": "2014-spring",
    "season_name_text": "2014年春",
    "released_on": "2014-04-10",
    "released_on_about": "",
    "official_site_url": "http://www.gochiusa.com/",
    "wikipedia_url": "http://ja.wikipedia.org/wiki/%E3%81%94%E6%B3%A8%E6%96%87%E3%81%AF%E3%81%86%E3%81%95%E3%81%8E%E3%81%A7%E3%81%99%E3%81%8B%3F#.E3.83.86.E3.83.AC.E3.83.93.E3.82.A2.E3.83.8B.E3.83.A1",
    "twitter_username": "usagi_anime",
    "twitter_hashtag": "gochiusa",
    "episodes_count": 12,
    "watchers_count": 850
  },
  "episode": {
    "id": 5013,
    "number": null,
    "number_text": "第1羽",
    "sort_number": 1,
    "title": "ひと目で尋常でないもふもふだと見抜いたよ",
    "records_count": 103
  }
}
```


# DELETE /v1/me/records/:id

作成した記録を削除することができます。

**このリクエストには `write` スコープが必要になります。**

## パラメータ

| 名前 | 概要 | データ例 |
| --- | --- | --- |
| id | **[必須]** 記録ID | 1016 |


## リクエスト例

```
$ curl -X DELETE https://api.annict.com/v1/me/records/1016?access_token=(access_token)
```

```
HTTP/1.1 204 No Content
```
