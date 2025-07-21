# typed: false
# frozen_string_literal: true

FactoryBot.define do
  factory :forum_category do
    sequence(:slug) { |n| [:site_news, :general, :feedback, :db_request][n % 4] }
    sequence(:name) { |n| "フォーラムカテゴリ#{n}" }
    sequence(:name_en) { |n| "Forum Category #{n}" }
    sequence(:description) { |n| "フォーラムカテゴリ#{n}の説明" }
    sequence(:description_en) { |n| "Description for Forum Category #{n}" }
    postable_role { "user" }
    sequence(:sort_number) { |n| n }
    forum_posts_count { 0 }

    trait :site_news do
      slug { :site_news }
      name { "サイトのお知らせ" }
      name_en { "Site News" }
      description { "Annictからのお知らせ" }
      description_en { "Announcements from Annict" }
      postable_role { "admin" }
    end

    trait :general do
      slug { :general }
      name { "雑談" }
      name_en { "General" }
      description { "Annictに関する雑談" }
      description_en { "General discussions about Annict" }
    end

    trait :feedback do
      slug { :feedback }
      name { "フィードバック" }
      name_en { "Feedback" }
      description { "Annictへのフィードバック" }
      description_en { "Feedback for Annict" }
    end

    trait :db_request do
      slug { :db_request }
      name { "データベースリクエスト" }
      name_en { "Database Request" }
      description { "作品情報の追加・修正リクエスト" }
      description_en { "Requests for adding or modifying work information" }
    end
  end
end
