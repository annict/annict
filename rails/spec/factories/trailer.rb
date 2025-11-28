# typed: false
# frozen_string_literal: true

FactoryBot.define do
  factory :trailer do
    work
    url { "https://www.youtube.com/watch?v=2ZR6fCnPcvA" }
    title { "コミックマーケット86公開PV" }

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
