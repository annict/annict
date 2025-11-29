# typed: false
# frozen_string_literal: true

FactoryBot.define do
  factory :userland_project_member do
    association :user
    association :userland_project
  end
end
