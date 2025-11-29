# typed: false
# frozen_string_literal: true

FactoryBot.define do
  factory :setting do
    association :user
    privacy_policy_agreed { true }
  end
end
