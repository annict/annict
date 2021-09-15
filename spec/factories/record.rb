# frozen_string_literal: true

FactoryBot.define do
  factory :record do
    association :user, :with_profile
    association :work
    body { "おもしろかった" }
    rating { "good" }

    trait :for_episode do
      association :episode
      association :recordable, factory: :episode_record
    end

    trait :for_work do
      association :recordable, factory: :work_record
    end
  end
end
