# frozen_string_literal: true

FactoryBot.define do
  factory :staff do
    association :resource, factory: :person
    work
    sequence(:name) { |n| "山田#{n}郎" }
    sequence(:name_en) { |n| "Yamada, #{n}rou" }
    role { :original_creator }
    role_other { "role_other_data" }
    role_other_en { "role_other_en_data" }

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
