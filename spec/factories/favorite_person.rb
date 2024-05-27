# typed: false
# frozen_string_literal: true

FactoryBot.define do
  factory :person_favorite do
    person
    user
  end
end
