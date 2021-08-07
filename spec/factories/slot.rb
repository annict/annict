# frozen_string_literal: true

FactoryBot.define do
  factory :slot do
    association :program
    association :episode
    anime { episode.anime }
    channel { program.channel }
    started_at { Time.parse("2017-01-29 00:00:00").utc }
    rebroadcast { false }

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
