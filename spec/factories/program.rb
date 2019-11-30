# frozen_string_literal: true

FactoryBot.define do
  factory :program do
    association :channel
    association :work
    started_at { Time.parse("2017-01-29 0:00:00") }
    rebroadcast { false }
  end
end
