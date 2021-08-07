# frozen_string_literal: true

FactoryBot.define do
  factory :cast do
    person
    anime
    character
    sequence(:name) { |n| "山田#{n}郎" }
    sequence(:name_en) { |n| "Yamada, #{n}rou" }

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
