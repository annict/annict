# frozen_string_literal: true

FactoryBot.define do
  factory :episode do
    anime
    sequence(:number) { |n| "第#{n}話" }
    sequence(:title) { |n| "Yes! プリキュア#{n}" }

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
