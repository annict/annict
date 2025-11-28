# typed: false
# frozen_string_literal: true

FactoryBot.define do
  factory :channel_work do
    user
    work
    channel
  end
end
