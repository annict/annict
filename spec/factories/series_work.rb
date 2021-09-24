# frozen_string_literal: true

FactoryBot.define do
  factory :series_work do
    series
    work
    summary { "TVシリーズ" }

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
