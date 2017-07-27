# frozen_string_literal: true

namespace :tmp do
  task add_faqs: :environment do
    ActiveRecord::Base.transaction do
      [
        {
          category: {
            name: "アニメデータ",
            locale: "ja",
            sort_number: 10_000
          },
          contents: [
            {
              question: "作品データの不備はどこから報告したら良いですか？",
              answer: <<~EOS,
                <a href="/forum">フォーラム</a>や<a href="https://twitter.com/AnnictJP" target="_blank">Twitter</a>からご連絡頂ければ幸いです。
              EOS
              locale: "ja",
              sort_number: 10_000
            },
            {
              question: "製作委員会がものによって団体に登録されていたり人物に登録されていたりするようです。正しくは団体でしょうか？",
              answer: <<~EOS,
                正しくは団体です。
                スタッフのデータを外部から一括でインポートしたときの仕様で、インポート時の団体データが人物データとして登録されています。こちらは気が向いたときに修正しています。
              EOS
              locale: "ja",
              sort_number: 20_000
            }
          ]
        },
        {
          category: {
            name: "視聴ステータス",
            locale: "ja",
            sort_number: 20_000
          },
          contents: [
            {
              question: "「中断」と「中止」の違いは何ですか？",
              answer: <<~EOS,
                以下のような想定で中断と中止ステータスが存在します。

                - 中断: 途中まで見たけど続きはしばらくしたら見よう、という作品に設定するステータス
                - 中止: 途中まで見たけど続きはもう見る気が無い、という作品に設定するステータス

                ちなみに「見たい」は、まだ1話も見ていないけど見る気がある、という作品に設定される想定です。
              EOS
              locale: "ja",
              sort_number: 10_000
            }
          ]
        },
        {
          category: {
            name: "放送予定",
            locale: "ja",
            sort_number: 30_000
          },
          contents: [
            {
              question: "放送予定に「見てる」や「見たい」にしている作品が出てこないのはなぜ？",
              answer: <<~EOS,
                作品ごとのチャンネルの設定がされていない可能性があります。<a href="/channel/works">こちら</a>から設定内容が確認できます。
              EOS
              locale: "ja",
              sort_number: 10_000
            }
          ]
        },
        {
          category: {
            name: "Twitter/Facebook",
            locale: "ja",
            sort_number: 40_000
          },
          contents: [
            {
              question: "記録やレビューがTwitterにシェアできなくなったのですが？",
              answer: <<~EOS,
                TwitterからAnnictのアプリが削除された可能性があります。AnnictのTwitterアプリを削除すると、Twitter API経由でツイートできなくなります。
                Twitterのアプリは<a href="https://twitter.com/settings/applications" target="_blank">こちら</a>から確認することができます。
              EOS
              locale: "ja",
              sort_number: 10_000
            }
          ]
        },
        {
          category: {
            name: "機能追加・要望",
            locale: "ja",
            sort_number: 50_000
          },
          contents: [
            {
              question: "要望はどこから伝えれば良いですか？",
              answer: <<~EOS,
                <a href="/forum">フォーラム</a>や<a href="https://twitter.com/AnnictJP" target="_blank">Twitter</a>からご連絡頂ければ幸いです。
              EOS
              locale: "ja",
              sort_number: 10_000
            },
            {
              question: "利用者の「見た」を学習してリコメンドする機能は作らないの？",
              answer: <<~EOS,
                昔あったんですが、精度が悪くて消した過去があります。精度が高めのものをもう一度作りたいという気持ちはあります。
              EOS
              locale: "ja",
              sort_number: 20_000
            }
          ]
        },
        {
          category: {
            name: "その他",
            locale: "ja",
            sort_number: 60_000
          },
          contents: [
            {
              question: "対応ブラウザは何ですか？",
              answer: <<~EOS,
                対応ブラウザは以下になります。

                - PC: Chrome, Firefox, Safariの最新版
                - iOS: Safari, Chromeの最新版
                - Android: Chromeの最新版
              EOS
              locale: "ja",
              sort_number: 10_000
            },
            {
              question: "「Annict」ってどういう意味ですか？",
              answer: <<~EOS,
                「Anime」と「Addict (中毒者・依存者)」をかけ合わせた造語になります。
                「アニメを見ない生活はあり得ない！」と思っている人のためのサービスにしたいという気持ちから名付けました。
              EOS
              locale: "ja",
              sort_number: 20_000
            }
          ]
        }
      ].each do |data|
        category = FaqCategory.new(data[:category])
        category.save!
        puts "category #{category.name} created"

        data[:contents].each do |content_data|
          content = category.faq_contents.new(content_data)
          content.save!
          puts "category #{category.name} content #{content.question} created"
        end
      end
    end
  end
end
