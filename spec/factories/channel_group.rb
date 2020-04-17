# frozen_string_literal: true

FactoryBot.define do
  factory :channel_group do
    sequence(:sc_chgid)
    name { "テレビ 関東" }

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
