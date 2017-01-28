# POST /v1/me/statuses

作品のステータスを設定することができます。

**このリクエストには `write` スコープが必要になります。**

## パラメータ

| 名前 | 概要 | データ例 |
| --- | --- | --- |
| work_id | **[必須]** 作品ID | 1234 |
| kind | **[必須]** ステータスの種類。`wanna_watch` (見たい), `watching` (見てる), `watched` (見た), `on_hold` (中断), `stop_watching` (中止), または `no_select` (未選択) が指定できます。 | wanna_watch |


## リクエスト例

```
$ curl -X POST https://api.annict.com/v1/me/statuses?work_id=438&kind=watching&access_token=(access_token)
```

```
HTTP/1.1 204 No Content
```
