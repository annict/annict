# typed: false
# frozen_string_literal: true

FactoryBot.define do
  factory :forum_post do
    association :user
    association :forum_category
    sequence(:title) { |n| "フォーラム投稿タイトル#{n}" }
    sequence(:body) { |n| "フォーラム投稿の本文#{n}\n\nこれはテスト用の投稿です。" }
    forum_comments_count { 0 }
    edited_at { nil }
    last_commented_at { Time.current }
    locale { "ja" }

    trait :in_english do
      locale { "en" }
      sequence(:title) { |n| "Forum Post Title #{n}" }
      sequence(:body) { |n| "Forum post body #{n}\n\nThis is a test post." }
    end

    trait :edited do
      edited_at { Time.current }
    end

    trait :with_comments do
      forum_comments_count { 3 }
      last_commented_at { 1.hour.ago }
    end

    trait :site_news do
      association :forum_category, :site_news
      sequence(:title) { |n| "お知らせ: アップデート v#{n}.0.0" }
      sequence(:body) { |n| "本日、Annict v#{n}.0.0をリリースしました。\n\n変更内容:\n- 新機能の追加\n- バグ修正" }
    end

    trait :general_discussion do
      association :forum_category, :general
      sequence(:title) { |n| "雑談トピック#{n}" }
      sequence(:body) { |n| "みなさん、こんにちは！\n\n雑談トピック#{n}です。" }
    end

    trait :feedback_post do
      association :forum_category, :feedback
      sequence(:title) { |n| "機能リクエスト: #{n}" }
      sequence(:body) { |n| "こんな機能があったらいいなと思います。\n\n詳細: ..." }
    end

    trait :db_request_post do
      association :forum_category, :db_request
      sequence(:title) { |n| "作品追加リクエスト: アニメタイトル#{n}" }
      sequence(:body) { |n| "以下の作品の追加をお願いします。\n\nタイトル: アニメタイトル#{n}\n放送時期: 2024年春" }
    end
  end
end
