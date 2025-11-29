# typed: false
# frozen_string_literal: true

FactoryBot.define do
  factory :person do
    sequence(:name) { |n| "山田#{n}郎" }
    sequence(:name_kana) { |n| "やまだ#{n}ろう" }
    sequence(:name_en) { |n| "Yamada, #{n}rou" }
    nickname { "山田" }
    nickname_en { "Yamada" }
    gender { :male }
    url { "http://example.com" }
    url_en { "http://example.com?lang=en" }
    wikipedia_url { "https://ja.wikipedia.org" }
    wikipedia_url_en { "https://wikipedia.org" }
    twitter_username { "example" }
    twitter_username_en { "example_en" }
    birthday { Date.parse("2000-01-01") }
    blood_type { :a }
    height { 150 }

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
