# ベースURI

Annict APIのベースURIは下記になります。

```
https://api.annict.com
```

Annictが提供する全てのAPIには、このベースURIの後にエンドポイントのパスを記述してアクセスすることができます。例えば作品情報を取得する場合は、パス `/v1/works` を繋げて下記のようにリクエストを送ります。

```
GET https://api.annict.com/v1/works
```


# 日付

フォーマットは[ISO 8601](http://www.iso.org/iso/home/standards/iso8601.htm) (`YYYY-MM-DDTHH:MM:SSZ`) で返します。よってタイムゾーンはUTCとなります。


# 認可 / 認証方法

Annict APIではリソースへの認可や認証に「OAuth 2.0」を使用しています。OAuth 2.0については [RFC 6749](https://tools.ietf.org/html/rfc6749) をご覧ください。


# アクセストークンの付与

Annict APIを通じてリソースを取得するためには、OAuthプロバイダ (Annict) が発行したアクセストークンをリクエストに付与する必要があります。付与の仕方は2種類あります。

1. リクエストヘッダに付与する
2. URIのパラメータに付与する


## リクエスト例 (リクエストヘッダに付与する)

```
$ curl -H "Authorization: Bearer 35372b2d866222ed33e355c36d86be498076e037a810ee72963819339c781f32" \
-X GET http://api.annict.com/v1/works
```


## リクエスト例 (URIのパラメータに付与する)

```
$ curl -X GET http://api.annict.com/v1/works?access_token=35372b2d866222ed33e355c36d86be498076e037a810ee72963819339c781f32
```


## レスポンス例

```json
{
  "works": [
    {
      "id":1129,
      "title":"精霊の守り人",
      "title_kana":"せいれいのもりびと",
      "media":"tv",
      ...
    }
  ]
}
```

アクセストークンの取得方法は [認証](https://annict.wikihub.io/wiki/api/authentication) ページに記載しています。


# HTTPステータスコード

Annict APIでは下記の条件でHTTPステータスコードを返します。

| ステータスコード | 条件 |
| --- | --- |
| 200 (OK) | リクエストが成功し、同時に処理が正常に完了したとき。 |
| 201 (Created) | リクエストが成功し、新しいリソースが作られたとき。 |
| 202 (Accepted) | リクエストは成功したが、処理は非同期で実行されるとき。 |
| 204 (No Content) | リクエストが成功し、レスポンスボディが存在しないとき。`DELETE` メソッドでリソースが削除されたときに使用します。 |
| 400 (Bad Request) | リクエストの内容が正しくないとき。 |
| 401 (Unauthorized) | リソースへのアクセスに認証が必要なとき。リクエストを送ったクライアントが特定できないときに使用します。 |
| 403 (Forbidden) | リソースへのアクセスが禁止されているとき。リクエストを送ったクライアントは特定したものの、指定されたリソースへのアクセスを禁止しているときに使用します。 |
| 404 (Not Found) | 指定されたリソースが存在しないとき。 |
| 500 (Internal Server Error) | Annictのサーバ内部に問題があり正常にレスポンスが返せないとき。 |
| 503 (Service Unavailable) | Annictのサーバが一時的に停止しているとき。 |


# ページネーション

作品一覧など、複数のリソースを返すAPIにはページを指定して取得するリソースを分割するようになっています。ページを指定する場合は `page` パラメータを、1ページあたり何件のリソースを取得するかを指定する場合は `per_page` パラメータを使用します。デフォルト値などの情報は各エンドポイントのドキュメントに記載しています。
レスポンスデータには `total_count`, `next_page`, `prev_page` の3つのフィールドが付加されます。`total_count` にはページを跨いだ全リソース数、`next_page`, `prev_page` には次のページ数、前のページ数が格納されています。次のページ、前のページが存在しない場合は `null` が格納されます。
以下は全8件の作品情報の4ページ目を2件取得する例です。5ページ目は存在しないため、`next_page` が `null` になっています。

```
GET https://api.annict.com/v1/works?per_page=2&page=4&access_token=(access_token)
```

```json
{
  "works": [
    ...
  ],
  "total_count": 8,
  "next_page": null,
  "prev_page": 3
}
```


# エラー

Annict APIがエラーを返すときは、基本的に以下の形式でレスポンスボディを返します。

```json
"errors": [{
  "type": "invalid_params",
  "message": "リクエストに失敗しました",
  "developer_message": "per_pageは1以上の値にしてください",
  "url": "http://example.com/docs/api/validations"
}]
```

| キー名 | 概要 |
| --- | --- |
| type | 発生したエラーのタイプ。全てのタイプは下記の表にまとめています。 |
| message | アプリケーションの利用者向けのエラーメッセージ。 |
| developer_message | APIを利用している開発者向けのエラーメッセージ。 |

| type | 概要 |
| --- | --- |
| invalid_params | リクエスト時に渡したパラメータに不備がある。 |
