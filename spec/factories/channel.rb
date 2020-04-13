# frozen_string_literal: true

FactoryBot.define do
  factory :channel do
    association :channel_group
    sequence :sc_chid
    name { "テレビ夕日" }

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
