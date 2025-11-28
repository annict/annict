# typed: false
# frozen_string_literal: true

FactoryBot.define do
  factory :channel do
    association :channel_group
    sequence(:sc_chid) { |n| 100000 + n }
    sequence(:name) { |n| "チャンネル#{n}" }
    sort_number { 100 }
    vod { false }

    trait :with_vod do
      vod { true }
    end
  end
end
