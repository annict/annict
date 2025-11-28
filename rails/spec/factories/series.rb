# typed: false
# frozen_string_literal: true

FactoryBot.define do
  factory :series do
    sequence(:name) { |n| "#{n}人はプリキュア" }
    sequence(:name_ro) { |n| "#{n} ha Precure" }
    sequence(:name_en) { |n| "#{n} ha Precure" }

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
