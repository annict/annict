# frozen_string_literal: true

FactoryGirl.define do
  factory :channel_work do
    user
    work
    channel
  end
end
