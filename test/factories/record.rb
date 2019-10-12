# frozen_string_literal: true

FactoryBot.define do
  factory :record do
    association :user, :with_profile
    association :work
  end
end
