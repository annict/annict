# frozen_string_literal: true

FactoryBot.define do
  factory :work do
    sequence(:title) { |n| "#{n}人はプリキュア" }
    sequence(:title_kana) { |n| "#{n}にんはぷりきゅあ" }
    media { :tv }
    official_site_url { "http://example.com" }
    wikipedia_url { "http://wikipedia.org" }
    synopsis { "プリキュアのあらすじ" }
    synopsis_source { "あらすじのソース" }
    twitter_username { "precure_official" }
    twitter_hashtag { "precure" }
    mal_anime_id { 12_345 }
    released_at { Date.parse("2012-04-05") }
    released_at_about { "2012年" }

    trait :with_current_season do
      year, name = ENV["ANNICT_CURRENT_SEASON"].split("-")
      season_year { year }
      season_name { name }
    end

    trait :with_next_season do
      year, name = ENV["ANNICT_NEXT_SEASON"].split("-")
      season_year { year }
      season_name { name }
    end

    trait :with_prev_season do
      year, name = ENV["ANNICT_PREVIOUS_SEASON"].split("-")
      season_year { year }
      season_name { name }
    end

    trait :with_episode do
      after :create do |work|
        create(:episode, work: work)
      end
    end
  end
end
