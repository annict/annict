# frozen_string_literal: true

FactoryBot.define do
  factory :record do
    association :user, :with_profile
    association :work
    watched_at { Time.zone.now }

    trait :with_episode_record do
      transient do
        episode { nil }
      end

      after(:create) do |record, evaluator|
        episode = evaluator.episode.presence || create(:episode, work: record.work)
        create :episode_record, user: record.user, record: record, episode: episode
      end
    end

    trait :with_work_record do
      after(:create) do |record|
        create :work_record, user: record.user, record: record, work: record.work
      end
    end
  end
end
