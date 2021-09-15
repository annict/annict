# frozen_string_literal: true

FactoryBot.define do
  factory :like do
    association :user, :with_profile
    association :likeable
  end
end
