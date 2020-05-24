# frozen_string_literal: true

FactoryBot.define do
  factory :episode_record do
    association :user, :with_profile
    body { "おもしろかった" }
    twitter_url_hash { |n| "xxxxx#{n}" }
    episode
    rating { 3.0 }

    before(:create) do |episode_record|
      episode_record.work = episode_record.episode.work
      episode_record.record = create(:record, user: episode_record.user, work: episode_record.work)
    end
  end
end
