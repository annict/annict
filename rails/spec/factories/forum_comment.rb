# typed: false
# frozen_string_literal: true

FactoryBot.define do
  factory :forum_comment do
    association :forum_post
    association :user
    sequence(:body) { |n| "これはコメント#{n}です。\n\nフォーラム投稿に対する返信です。" }
    edited_at { nil }
    locale { "ja" }

    trait :in_english do
      locale { "en" }
      sequence(:body) { |n| "This is comment #{n}.\n\nA reply to the forum post." }
    end

    trait :edited do
      edited_at { Time.current }
    end

    trait :long do
      body { "これは長いコメントです。\n\n" + ("テキスト" * 100) }
    end

    trait :with_mention do
      sequence(:body) { |n| "@user_name さん、こんにちは！\n\nコメント#{n}です。" }
    end
  end
end
