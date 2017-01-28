Annict APIからリソースにアクセスするには、OAuth 2による認可と認証を行う必要があります。

# クライアントアプリケーションを作成する

OAuthで使用するクライアントIDなどを保持するクライアントアプリケーションを作成します。クライアントアプリケーションは https://annict.com/oauth/applications から作成することができます。

# 認可のリクエストを送る

Annict APIが提供するリソースにどのような権限でアクセスするかをAnnictに伝えるため、認可のためのリクエストを送る必要があります。

## GET /oauth/authorize

指定した権限でリソースにアクセスすることの許可を求めるページが表示されます。

### パラメータ

| 名前 | 概要 |
| --- | --- |
| client_id | **[必須]** クライアントID。作成したクライアントアプリケーションのclient_idになります。 |
| response_type | **[必須]** 認可後にコールバックで返ってきたときのURIに付与される認証コードのパラメータ名。 `code` を指定してください。 |
| redirect_uri | **[必須]** 認可後のコールバックでどのURIに返るかを指定します。`urn:ietf:wg:oauth:2.0:oob` を指定した場合、認可後にリダイレクトせず認証コードが表示されます。 |
| scope | リソースにアクセスする際の権限を指定します。デフォルトは `read` (読み込み専用) になっています。書き込み権限も付与したい場合は `read write` を指定してください。 |

### リクエスト例

下記の例では `redirect_uri` に `urn:ietf:wg:oauth:2.0:oob` が指定されているため、認可後のページに認証コードが表示されます。

```
GET /oauth/authorize?client_id=f96b162be2c54b467f7583a0e141d0df99bf16318146f9bf53116d82b145fde6&response_type=code&redirect_uri=urn%3Aietf%3Awg%3Aoauth%3A2.0%3Aoob&scope=read+write
```

下記の例では `redirect_uri` に `http://example.com` が指定されているため、認可後に `http://example.com/?code=xxx` というページにリダイレクトされます。パラメータ `code` に渡されている値が認証コードになります。

```
GET /oauth/authorize?client_id=f96b162be2c54b467f7583a0e141d0df99bf16318146f9bf53116d82b145fde6&response_type=code&redirect_uri=http://example.com&scope=read+write
```

# アクセストークンを取得する

認可後に取得した認証コードを使用してアクセストークンを取得します。

## POST /oauth/token

アクセストークンを発行します。

### パラメータ

| 名前 | 概要 |
| --- | --- |
| client_id | **[必須]** クライアントID。 |
| client_secret | **[必須]** 作成したクライアントアプリケーションのシークレットキー。 |
| grant_type | **[必須]** `authorization_code` を指定してください。 |
| redirect_uri | **[必須]** クライアントアプリケーションを作成したときに入力したコールバックURIを指定します。 |
| code | **[必須]** 認可後に取得した認証コードを指定します。 |

### リクエスト例

```
$ curl -F client_id=f96b162be2c54b467f7583a0e141d0df99bf16318146f9bf53116d82b145fde6 \
-F client_secret=9bef7fc39d5d6210ecaaa34d8ccfe5e7a10babd2ebad90876307d71a8e14de69 \
-F grant_type=authorization_code \
-F redirect_uri=http://example.com \
-F code=82c0436de2b586ca974b45ab4e12ceeeeda2a8cf90a269b2dbc5a24c4a6fc8f2 \
-X POST https://api.annict.com/oauth/token
```

### レスポンス例

```
HTTP/1.1 200 OK
ontent-Type: application/json; charset=utf-8
```

```json
{
  "access_token":"58468586b6f4c29e88a8d9b7f3babb8364fec9991f2681081cbfd849d7c11a91",
  "token_type":"bearer",
  "scope":"read write",
  "created_at":1465718311
}
```

# アクセストークンの情報を取得する

## GET /oauth/token/info

アクセストークンの情報を取得します。`Authorization` ヘッダにアクセストークンを渡してリクエストを送ります。

### リクエスト例

```
$ curl -H "Authorization: Bearer 35372b2d866222ed33e355c36d86be498076e037a810ee72963819339c781f32" \
-X GET https://api.annict.com/oauth/token/info
```

### レスポンス例

```
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
```

```json
{
  "resource_owner_id":2,
  "scopes":["read","write"],
  "expires_in_seconds":null,
  "application": {
    "uid":"bfa6272ede7f7d096099c4e5f15923a78630520435169f47006e18d5e4dc2a7e"
  },
  "created_at":1461949248
}
```

# アクセストークンを失効させる

## POST /oauth/revoke

アクセストークンを失効させます。`Authorization` ヘッダと `token` パラメータを渡してリクエストを送ります。

### パラメータ

| 名前 | 概要 |
| --- | --- |
| token | **[必須]** アクセストークンを指定します。 |

### リクエスト例

```
$ curl -H "Authorization: Bearer 35372b2d866222ed33e355c36d86be498076e037a810ee72963819339c781f32" \
-F token=35372b2d866222ed33e355c36d86be498076e037a810ee72963819339c781f32
-X POST https://api.annict.com/oauth/revoke
```

### レスポンス例

```
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
```

```json
{}
```
