# Annict

[![Travis CI](https://travis-ci.org/annict/annict.svg?branch=master)](https://travis-ci.org/annict/annict)
[![Code Climate](https://codeclimate.com/github/annict/annict/badges/gpa.svg)](https://codeclimate.com/github/annict/annict)
[![Coveralls](https://coveralls.io/repos/github/annict/annict/badge.svg?branch=master)](https://coveralls.io/github/annict/annict?branch=master)
[![Gemnasium](https://gemnasium.com/annict/annict.svg)](https://gemnasium.com/annict/annict)
[![Slack](http://slack.annict.com/badge.svg)](http://slack.annict.com)
[![Waffle](https://badge.waffle.io/annict/annict.svg?label=TODO&title=TODO)](http://waffle.io/annict/annict)


### 開発に参加する

#### 新機能・改善案の話やバグの話など

開発に関するもろもろのやり取りは、以下のサービスで行っています。

* [Slack](http://slack.annict.com)
* [Twitter](https://twitter.com)
  * [@anannict](https://twitter.com/anannict) へのメンションなどに反応します


#### セキュリティに関わるバグの報告

admin@annict.com までメールをください。
Twitterなど、第三者に公開されている場所に投稿しないでもらえるとありがたいです。


#### Pull Requests

絶賛募集中です！以下の項目を守った上で送ってもらえると嬉しいです。

* [コーディング規約](https://github.com/annict/annict/wiki/%E3%82%B3%E3%83%BC%E3%83%87%E3%82%A3%E3%83%B3%E3%82%B0%E8%A6%8F%E7%B4%84)を意識したコードを書いてください
* 新たに機能を追加したときはそのテストも追加してください
* 既存のテストを全てパスすることを確認してください


#### タスク管理について

Annictでは開発に関係するタスク管理を[Waffle](https://waffle.io/annict/annict)で行っています。
各タスクは以下のリストで管理しています。

| リスト名 | 概要 |
| ------- | ----------- |
| Just An Idea | システム内に取り込むことが決定していない漠然としたタスク |
| TODO   | システム内に取り込むことが決定したタスク |
| In Progress | 現在取り組んでいるタスク |
| Done    | 作業が完了したタスク |

「TODO」リストの上から順に優先度が高いタスクとなっています。


#### 開発環境を作る

##### 必要なものをインストールする

Annictは以下のソフトウェアを使用して開発しています。
Annictを動かすには事前にこれらをインストールする必要があります。

* Ruby 2.3
* PostgreSQL 9.5
* ImageMagick
* Node.js 0.12
* PhantomJS
  * テストを実行するときに使用しています。Annictをローカルで動かすだけであれば不要です


##### Annictを動かす

GitHubからソースコードをcloneしてから以下のコマンドを実行してください。

```
$ cd annict
$ cp config/application.yml{.example,}
$ bundle
$ rake db:setup
$ npm install
$ rails s -b 0.0.0.0
```

[http://localhost:3000](http://localhost:3000) にアクセスすると、
サイトのトップページが表示されるはずです。

ただ、上記コマンドを実行しただけでは作品画像などは表示されないはずです。
画像を表示するには「Tombo」という画像変換サーバを別途起動する必要があります。


##### 画像変換サーバ「Tombo」について

Annictでは作品やアバターなど画像を表示するとき、
「Tombo」という画像を動的にリサイズする画像変換サーバを使用しています。
動かし方などは[Tomboのリポジトリ](https://github.com/shimbaco/tombo)をご覧ください。

AnnictでTomboを使用するときは `localhost:5000` でサーバを起動します。


##### application.ymlを編集する

AnnictのRailsアプリに必要な設定値は全て `config/application.yml` に記述しています。
開発環境でも必要に応じて設定を変更する必要があります。

例えばTwitterアカウントを使用してローカル環境でユーザ登録をしたいときは、
TwitterでOAuth認証用のアプリを作成し、`config/application.yml` に記述されている
`TWITTER_CONSUMER_KEY` と `TWITTER_CONSUMER_SECRET` を変更します。


#### テストを実行する

以下のコマンドでテストが実行できます。

```
$ rspec
```


### ライセンス

Copyright 2014-2017 Annict

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
