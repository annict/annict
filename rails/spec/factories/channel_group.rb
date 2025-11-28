# typed: false
# frozen_string_literal: true

FactoryBot.define do
  factory :channel_group do
    sequence(:name) { |n| "チャンネルグループ#{n}" }
    sequence(:sort_number) { |n| n }

    trait :unpublished do
      unpublished_at { Time.current }
    end

    trait :deleted do
      deleted_at { Time.current }
    end
  end
end
