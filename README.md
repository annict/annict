# Annict

[![Circle CI](https://circleci.com/gh/annict/annict/tree/master.svg?style=svg)](https://circleci.com/gh/annict/annict/tree/master)  [![Code Climate](https://codeclimate.com/github/annict/annict/badges/gpa.svg)](https://codeclimate.com/github/annict/annict) [![Test Coverage](https://codeclimate.com/github/annict/annict/badges/coverage.svg)](https://codeclimate.com/github/annict/annict) [![Dependency Status](https://gemnasium.com/annict/annict.svg)](https://gemnasium.com/annict/annict)  [![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/annict/annict)

[![木崎湖](http://d3a8d1smk6xli.cloudfront.net/github/kizakiko.png)](http://ja.wikipedia.org/wiki/%E6%9C%A8%E5%B4%8E%E6%B9%96)

http://www.annict.com

見たアニメを記録するWebサービスです。


### スクショ

[![](http://d3a8d1smk6xli.cloudfront.net/github/screenshot3.gif)](http://d3a8d1smk6xli.cloudfront.net/github/screenshot3.gif)


### 開発に参加する

#### 新機能・改善案の話やバグの話など

開発に関わるもろもろのやり取りは、以下のサービスで行っています。

* [GitHub Issues](https://github.com/annict/annict)
* [Gitter](https://gitter.im/annict/annict)


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

** :warning: ここに書かれている情報は古いです。開発のほうが落ち着いたら更新しますm(__)m (2015年3月5日現在) **

##### 依存関係

Annictは以下のソフトウェアを使用して開発しています。事前にこれらをインストールしてください。

* Ruby 2.1.5
* PostgreSQL 9.3.5.0
* Redis 2.8.3
* ImageMagick 6.7
* PhantomJS 1.9
  * テストの実行時にしています。Annictをローカルで動かすだけであれば不要です


##### Annictを動かす

GitHubからソースコードをcloneしてから以下のコマンドを実行してください。

```
$ cd annict
$ cp config/application.yml{.example,}
$ bundle install
$ bundle exec rake db:create
$ bundle exec rake db:setup
$ foreman start
```

http://localhost:5000 にアクセスすると、サイトのトップページが表示されるはずです。


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
