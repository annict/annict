# GET /v1/episodes

Annictに登録されているエピソード情報を取得することができます。

## フィールド

| 名前 | 概要 |
| --- | --- |
| id | エピソードのID |
| number | エピソードの話数 |
| number_text | エピソードの話数 (表記用) |
| sort_number | ソート用の番号。話数でソートすると正しく並べられないケースがあるため、このフィールドが存在します。 |
| title | サブタイトル |
| records_count | 記録数 |
| work | このエピソードが紐づく作品情報。取得できるフィールドは [Works](https://annict.wikihub.io/wiki/api/works) と同じです。 |
| prev_episode | このエピソードの前のエピソード情報。取得できるフィールドは同じです。 |
| next_episode | このエピソードの次のエピソード情報。取得できるフィールドは同じです。 |


## パラメータ

| 名前 | 概要 | 使用例 |
| --- | --- | --- |
| fields | レスポンスボディに含まれるデータのフィールドを絞り込みます。 | fields=id,title |
| filter_ids | エピソードをエピソードIDで絞り込みます。 | filter_ids=1,2,3 |
| filter_work_id | エピソードを作品IDで絞り込みます。 | filter_work_id=1111 |
| page | ページ数を指定します。 | page=2 |
| per_page | 1ページに何件取得するかを指定します。デフォルトは `25` 件で、`50` 件まで指定できます。 | per_page=30 |
| sort_id | エピソードをエピソードIDで並び替えます。`asc` または `desc` が指定できます。 | sort_id=desc |
| sort_sort_number | エピソードをソート用の番号で並び替えます。`asc` または `desc` が指定できます。 | sort_sort_number=desc |


## リクエスト例

```
$ curl -X GET https://api.annict.com/v1/episodes?access_token=(access_token)
```

```json
{
  "episodes": [
    {
      "id": 45,
      "number": null,
      "number_text": "第2話",
      "sort_number": 2,
      "title": "殺戮の夢幻迷宮",
      "records_count": 0,
      "work": {
        "id": 3831,
        "title": "NEWドリームハンター麗夢",
        "title_kana": "",
        "media": "ova",
        "media_text": "OVA",
        "season_name": "1990-autumn",
        "season_name_text": "1990年秋",
        "released_on": "1990-12-16",
        "released_on_about": "",
        "official_site_url": "",
        "wikipedia_url": "",
        "twitter_username": "",
        "twitter_hashtag": "",
        "episodes_count": 2,
        "watchers_count": 10
      },
      "prev_episode": {
        "id": 44,
        "number": null,
        "number_text": "第1話",
        "sort_number": 1,
        "title": " 夢の騎士達",
        "records_count": 0
      },
      "next_episode": null
    },
...
```
