# frozen_string_literal: true

FactoryBot.define do
  factory :record do
    association :user, :with_profile
    association :anime

    trait :with_episode_record do
      transient do
        episode { nil }
      end

      after(:create) do |record, evaluator|
        episode = evaluator.episode.presence || create(:episode, anime: record.anime)
        create :episode_record, user: record.user, record: record, episode: episode
      end
    end

    trait :with_anime_record do
      after(:create) do |record|
        create :anime_record, user: record.user, record: record, anime: record.anime
      end
    end
  end
end
