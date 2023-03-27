# 開発環境を作る

## システム構成

Annictは以下のようなシステム構成になっています。

- アプリケーション
  - Ruby on Rails
- データベース
  - PostgreSQL
  - Redis
- 画像変換サーバ
  - [imgproxy](https://imgproxy.net/)

開発環境を作るには、上記のシステム構成をローカルに再現する必要があります。

## 用意するもの

以下を手元の開発環境にインストールします。

- Ruby
  - 必要なバージョンは [.tool-versions](https://github.com/annict/annict/blob/main/.tool-versions) に記載されています
- Node.js
  - 必要なバージョンは [.tool-versions](https://github.com/annict/annict/blob/main/.tool-versions) に記載されています
- Yarn
  - 必要なバージョンは [package.json](https://github.com/annict/annict/blob/main/package.json) に記載されていますが、最近のものであれば動作に支障ないと思います

## 手順

### `annict.test` でアクセスできるようにする

まずはローカルで動かすサーバに `annict.test` というホスト名でアクセスできるようにするため、以下のように `/etc/hosts` を編集します。

```sh
$ sudo sh -c "echo '127.0.0.1  annict.test' >> /etc/hosts"
```

### ソースコードを取得する

ソースコードをcloneします。

```sh
$ git clone git@github.com:annict/annict.git
```

### Dockerを使ってデータベースなどを立ち上げる

PostgreSQLやRedisといったデータベースやimgproxyはDockerを使って立ち上げるようになっています。

`docker compose up` します。

```sh
$ cd /path/to/annict
$ docker compose up
```

### Railsのセットアップをする

`docker compose up` したターミナルはそのままに、別のターミナルを立ち上げて以下を実行します。

```sh
$ cd /path/to/annict

// パッケージのインストール
$ bundle install
$ yarn install

// データベースの初期化
$ bin/rails db:setup

// JS/CSSをコンパイルするプロセスを立ち上げる
$ foreman start -f Procfile.dev

// サーバを起動する
$ bin/rails s
```

[http://annict.test:3000](http://annict.test:3000) にアクセスすると、トップページが表示されるはずです。

### 管理者を作成する

開発用サーバを立ち上げただけの状態だとユーザやアニメのデータが登録されていないため、ほぼ何もできません。
[Annict DB](http://annict.test:3000/db)からアニメデータを登録するため、まずは管理者を作成します。

まず `rails console` します。

```sh
$ bin/rails console
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
$ bin/rspec
```

## 画像のアップロードや表示について

このドキュメントでは簡単のため画像のアップロードや表示については触れていません。
画像をアップロードしたり表示できるようにするようにするにはAmazon S3 (互換のストレージ) やimgproxyを追加でセットアップする必要があります。
興味ある方は他のドキュメント等を参考にセットアップしてください。
