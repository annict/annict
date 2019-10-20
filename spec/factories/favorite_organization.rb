# frozen_string_literal: true

FactoryBot.define do
  factory :favorite_organization do
    organization
    user
  end
end
