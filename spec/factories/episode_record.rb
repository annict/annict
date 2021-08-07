# frozen_string_literal: true

FactoryBot.define do
  factory :episode_record do
    association :user, :with_profile
    body { "おもしろかった" }
    twitter_url_hash { |n| "xxxxx#{n}" }
    anime
    episode
    record
    rating { 3.0 }
  end
end
