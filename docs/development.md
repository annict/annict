# 開発環境を作る

## 用意するもの

以下を手元の開発環境にインストールします。

- Docker Compose

## 手順

### `annict.test` でアクセスできるようにする

まずはローカルで動かすサーバに `annict.test` というホスト名でアクセスできるようにするため、以下のように `/etc/hosts` を編集します。

```sh
sudo sh -c "echo '127.0.0.1  annict.test' >> /etc/hosts"
```

### ソースコードを取得する

ソースコードをcloneします。

```sh
git clone git@github.com:annict/annict.git
```

### Docker Composeを使って各種サービスを立ち上げる

以下を実行します。

```sh
cd /path/to/annict
docker compose up
```

### MinIOとimgproxyのセットアップをする

開発環境で画像をアップロードするとき、ストレージとして[MinIO](https://github.com/minio/minio)を使っています。
また、アップロードした画像のリサイズなどをするために[imgproxy](https://imgproxy.net/)を使っています。
MinIOとimgproxyは `docker compose up` ですでに起動しているはずです。

http://localhost:9001/login にアクセスし、以下の情報でMinIOの管理コンソールにログインします。

- Username: `minio_admin`
- Password: `minio_admin`

ログイン後、 http://localhost:9001/access-keys にアクセスし、アクセスキーを作成します。
作成したアクセスキーとシークレットキーを以下に設定します。

- [.env.development](https://github.com/annict/annict/blob/main/.env.development) の `S3_ACCESS_KEY_ID` と `S3_SECRET_ACCESS_KEY`
- [.env.imgproxy](https://github.com/annict/annict/blob/main/.env.imgproxy) の `AWS_ACCESS_KEY_ID` と `AWS_SECRET_ACCESS_KEY`

次に http://localhost:9001/buckets にアクセスし、`annict-development` という名前のバケットを作成します。

### Railsのセットアップをする

以下を実行します。

```sh
cd /path/to/annict
docker compose exec app bin/setup
docker compose exec app bin/dev
docker compose exec app bin/rails jobs:work
docker compose exec app bin/rails server
```

### ブラウザでAnnictにアクセスする

[http://annict.test:3000](http://annict.test:3000) にアクセスすると、トップページが表示されるはずです。

### 管理者を作成する

開発用サーバを立ち上げただけの状態だとユーザやアニメのデータが登録されていないため、ほぼ何もできません。
[Annict DB](http://annict.test:3000/db)からアニメデータを登録するため、まずは管理者を作成します。

まず `rails console` します。

```sh
docker compose exec app bin/rails console
```

以下のスクリプトを実行して管理者を作成します。ユーザ名やメールアドレスなどは適宜置き換えてください。

```rb
user = User.new(username: "shimbaco", email: "me@shimba.co", password: "shimbaco", role: "admin", time_zone: "Asia/Tokyo", locale: "ja")
user.build_relations
user.save!
user.confirm
```

### アニメデータを登録する

[ログインページ](http://annict.test:3000/sign_in)からログイン後、[Annict DB](http://annict.test:3000/db)にアクセスしてアニメを登録します。

大量にデータを登録したい場合は `spec/factories/work.rb` などを参考に登録してください。

### テストを実行する

AnnictではRSpecを使ってテストを書いています。以下のコマンドでテストを実行することができます。

```sh
docker compose exec -e RAILS_ENV=test app bin/rspec
```
