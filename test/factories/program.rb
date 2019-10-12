# frozen_string_literal: true

FactoryBot.define do
  factory :program do
    association :program_detail
    association :work
    association :episode
    started_at { Time.parse("2017-01-29 0:00:00") }
    channel
  end
end
