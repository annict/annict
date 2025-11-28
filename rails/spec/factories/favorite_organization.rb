# typed: false
# frozen_string_literal: true

FactoryBot.define do
  factory :organization_favorite do
    organization
    user
  end
end
