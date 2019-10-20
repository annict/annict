# frozen_string_literal: true

FactoryBot.define do
  factory :favorite_person do
    person
    user
  end
end
