# typed: false
# frozen_string_literal: true

FactoryBot.define do
  factory :organization do
    sequence(:name) { |n| "A-#{n} Pictures" }
    sequence(:name_kana) { |n| "えー#{n}ぴくちゃーず" }
    sequence(:name_en) { |n| "A-#{n} Pictures" }
    sequence(:url) { |n| "http://a#{n}p.jp" }
    sequence(:url_en) { |n| "http://a#{n}p.jp?lang=en" }
    sequence(:wikipedia_url) { |n| "https://ja.wikipedia.org/wiki/A-#{n}_Pictures" }
    sequence(:wikipedia_url_en) { |n| "https://en.wikipedia.org/wiki/A-#{n}_Pictures" }
    sequence(:twitter_username) { |n| "a#{n}pictures" }
    sequence(:twitter_username_en) { |n| "a#{n}pictures_en" }

    trait :published do
      unpublished_at { nil }
    end

    trait :unpublished do
      unpublished_at { Time.zone.now }
    end

    trait :not_deleted do
      deleted_at { nil }
    end

    trait :deleted do
      deleted_at { Time.zone.now }
    end
  end
end
