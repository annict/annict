# 開発環境を作る

## Annict Web

### 開発用サーバを立ち上げる

まずはローカルで動かすサーバに `annict.test` と `annict-jp.test` でアクセスできるようにするため、以下のように `/etc/hosts` を編集します。
`annict.test` と `annict-jp.test` はそれぞれ本番環境の `annict.com` (英語圏向け) と `annict.jp` (日本語圏向け) に対応します。

```sh
$ sudo sh -c "echo '127.0.0.1  annict.test' >> /etc/hosts"
$ sudo sh -c "echo '127.0.0.1  annict-jp.test' >> /etc/hosts"
```

ソースコードをcloneします。

```sh
$ git clone git@github.com:kiraka/annict-web.git
```

`docker compose up` します。

```sh
$ cd annict-web
$ docker compose up
```

データベースの初期化を行います。

```sh
$ docker compose exec app bundle exec rails db:setup
```

[http://annict-jp.test:3000](http://annict-jp.test:3000) (または [http://annict.test:3000](http://annict.test:3000)) にアクセスすると、トップページが表示されるはずです。

### 管理者を作成する

開発用サーバを立ち上げただけの状態だとユーザやアニメのデータが登録されていないため、ほぼ何もできません。
[Annict DB](http://annict-jp.test:3000/db)からアニメデータを登録するため、まずは管理者を作成します。

まず `rails console` します。

```sh
$ docker compose exec app bundle exec rails console
```

以下のスクリプトを実行して管理者を作成します。ユーザ名やメールアドレスは適宜置き換えてください。

```rb
User.create!(username: 'shimbaco', email: 'me@shimba.co', role: 'admin', time_zone: 'Asia/Tokyo', locale: 'ja')
```

### アニメデータを登録する

[ログインページ](http://annict-jp.test:3000/sign_in)からログイン後、[Annict DB](http://annict-jp.test:3000/db)にアクセスしてアニメを登録します。

大量にデータを登録したい場合は `spec/factories/work.rb` などを参考に登録してください。

### テストを実行する

AnnictではRSpecを使ってテストを書いています。以下のコマンドでテストを実行することができます。

```sh
$ docker compose exec app bundle exec rails db:setup RAILS_ENV=test
$ docker compose exec app /bin/bash -c 'RAILS_ENV=test bundle exec rspec'
```
