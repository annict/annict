<p align="center">
  <a href="http://www.annict.com" target="_blank">
    <img src="http://d3a8d1smk6xli.cloudfront.net/github/annict-logo2.png" alt="Annict" width="150" height="150">
  </a>
  <br>
  <br>
  <img src="http://d3a8d1smk6xli.cloudfront.net/github/annict-text-logo.png" alt="Annict" width="240" height="60">
  <br>
  見たアニメを記録して、共有しよう<br>
  <a href="http://www.annict.com" target="_blank">http://www.annict.com</a>
  <br>
  <br>
  <a href="https://circleci.com/gh/annict/annict/tree/master"><img src="https://circleci.com/gh/annict/annict/tree/master.svg?style=svg"></a>
  <a href="https://codeclimate.com/github/annict/annict"><img src="https://codeclimate.com/github/annict/annict/badges/gpa.svg"></a>
  <a href="https://codeclimate.com/github/annict/annict"><img src="https://codeclimate.com/github/annict/annict/badges/coverage.svg"></a>
  <a href="https://gemnasium.com/annict/annict"><img src="https://gemnasium.com/annict/annict.svg"></a>
  <a href="http://slack.annict.com"><img src="http://slack.annict.com/badge.svg"></a>
</p>

---

### 開発に参加する

#### 新機能・改善案の話やバグの話など

開発に関わるもろもろのやり取りは、以下のサービスで行っています。

* [GitHub Issues](https://github.com/annict/annict/issues)
* [Slack](http://slack.annict.com)


#### セキュリティに関わるバグの報告

anannict@gmail.com までメールをください。
Twitterなど、第三者に公開されている場所に投稿しないでもらえるとありがたいです。


#### Pull Requests

絶賛募集中です！以下の項目を守った上で送ってもらえると嬉しいです。

* [コーディング規約](https://github.com/annict/annict/wiki/%E3%82%B3%E3%83%BC%E3%83%87%E3%82%A3%E3%83%B3%E3%82%B0%E8%A6%8F%E7%B4%84)を意識したコードを書いてください
* 新たに機能を追加したときはそのテストも追加してください
* 既存のテストを全てパスすることを確認してください


#### タスク管理について

Annictでは開発に関係するタスク管理を[Trello](https://trello.com/b/UinnA33N/annict)で行っています。
各タスクは以下のリストに入れて管理していて、上に置かれているタスクから順に対応しています。

| リスト名 | どんなタスクを入れているか |
| ------- | ----------- |
| Idea   | システム内に取り込むことが決定していない漠然としたアイデアなどを入れています |
| Todo   | システム内に取り込むことが決定したタスクを入れています |
| Doing | 現在取り組んでいるタスクを入れています |
| Done | 作業が完了したタスクを入れています |


#### 開発環境を作る

##### 依存関係

Annictは以下のソフトウェアを使用して開発しています。
Annictを動かすには事前にこれらをインストールする必要があります。

* Ruby 2.2.2
* PostgreSQL 9.3
* ImageMagick
* Node.js 0.12
* PhantomJS
  * テストの実行時にしています。Annictをローカルで動かすだけであれば不要です


##### Annictを動かす

GitHubからソースコードをcloneしてから以下のコマンドを実行してください。

```
$ cd annict
$ cp config/application.yml{.example,}
$ bundle install
$ bundle exec rake db:create
$ bundle exec rake db:setup
$ bundle exec rails s
```

[http://localhost:3000](http://localhost:3000) にアクセスすると、サイトのトップページが表示されるはずです。

##### 画像変換サーバ「Tombo」について

Annictでは作品やアバターなど画像を表示するとき、「Tombo」という動的にリサイズする画像変換サーバを使用しています。動かし方などは[Tomboのリポジトリ](https://github.com/shimbaco/tombo)をご覧ください。
AnnictでTomboを使用するときは `localhost:5000` でサーバを起動します。


##### application.ymlを編集する

AnnictのRailsアプリに必要な設定値は全て `config/application.yml` に記述しています。開発環境でも必要に応じて設定を変更する必要があります。

例えばTwitterアカウントを使用してローカル環境でユーザ登録をしたいときは、TwitterでOAuth認証用のアプリを作成し、`config/application.yml` の `TWITTER_CONSUMER_KEY` と `TWITTER_CONSUMER_SECRET` に値を設定します。


#### テストを実行する

以下のコマンドでテストが実行できます。

```
$ bundle exec rspec
```


### ライセンス

Copyright 2015 Annict

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
