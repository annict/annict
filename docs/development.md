# 開発環境を作る

## 用意するもの

以下を手元の開発環境にインストールします。

- Docker Compose

## 手順

### `annict.test` でアクセスできるようにする

ローカルで動かすサーバに `annict.test` というホスト名でアクセスできるようにするため、以下のように `/etc/hosts` を編集します。

```sh
sudo sh -c "echo '127.0.0.1  annict.test' >> /etc/hosts"
```

### ソースコードを取得する

ソースコードを clone します。

```sh
git clone git@github.com:annict/annict.git
```

### Docker Compose を使って各種サービスを立ち上げる

以下を実行します。

```sh
docker compose up
```

### MinIO と imgproxy のセットアップをする

開発環境で画像をアップロードするとき、ストレージとして[MinIO](https://github.com/minio/minio)を使っています。
また、アップロードした画像のリサイズなどをするために[imgproxy](https://imgproxy.net/)を使っています。
MinIO と imgproxy は `docker compose up` ですでに起動しているはずです。

http://localhost:19001/login にアクセスし、以下の情報で MinIO の管理コンソールにログインします。

- Username: `minio_admin`
- Password: `minio_admin`

次に「Create Bucket」をクリックして `annict-development` という名前のバケットを作成します。

### Rails のセットアップをする

以下を実行します。

```sh
cd /path/to/annict

# コンテナに入る
docker compose exec app zsh

bin/setup
bin/dev
bin/rails jobs:work
bin/rails server
```

### ブラウザで Annict にアクセスする

[http://annict.test:13000](http://annict.test:13000) にアクセスすると、トップページが表示されるはずです。

### 管理者を作成する

開発用サーバを立ち上げただけの状態だとユーザやアニメのデータが登録されていないため、ほぼ何もできません。
[Annict DB](http://annict.test:13000/db)からアニメデータを登録するため、まずは管理者を作成します。

まず `rails console` します。

```sh
bin/rails console
```

以下のスクリプトを実行して管理者を作成します。ユーザ名やメールアドレスなどは適宜置き換えてください。

```rb
user = User.new(username: "shimbaco", email: "me@shimba.co", password: "shimbaco", role: "admin", time_zone: "Asia/Tokyo", locale: "ja")
user.build_relations
user.save!
user.confirm
```

### アニメデータを登録する

[ログインページ](http://annict.test:13000/sign_in)からログイン後、[Annict DB](http://annict.test:13000/db)にアクセスしてアニメを登録します。

大量にデータを登録したい場合は `spec/factories/work.rb` などを参考に登録してください。

### テストを実行する

Annict では RSpec を使ってテストを書いています。以下のコマンドでテストを実行することができます。

```sh
bin/rspec
```
