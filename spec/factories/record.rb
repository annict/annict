# frozen_string_literal: true

FactoryBot.define do
  factory :record do
    association :user, :with_profile
    association :work
    body { "おもしろかった" }
    rating { "good" }
    watched_at { Time.zone.now }

    trait :on_episode do
      association :episode
      association :recordable, factory: :episode_record
    end

    trait :on_work do
      association :recordable, factory: :work_record
    end
  end
end
