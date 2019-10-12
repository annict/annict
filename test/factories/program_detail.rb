# frozen_string_literal: true

FactoryBot.define do
  factory :program_detail do
    association :channel
    association :work
    started_at { Time.parse("2017-01-29 0:00:00") }
  end
end
