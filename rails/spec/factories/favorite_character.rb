# typed: false
# frozen_string_literal: true

FactoryBot.define do
  factory :character_favorite do
    character
    user
  end
end
