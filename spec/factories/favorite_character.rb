# frozen_string_literal: true

FactoryBot.define do
  factory :favorite_character do
    character
    user
  end
end
