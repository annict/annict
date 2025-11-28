# typed: false
# frozen_string_literal: true

FactoryBot.define do
  factory :follow do
    association :user, :with_profile
    association :following, factory: :user
  end
end
