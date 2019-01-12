# frozen_string_literal: true

FactoryBot.define do
  factory :cast do
    person
    work
    character
    sequence(:name) { |n| "山田#{n}郎" }
    sequence(:name_en) { |n| "Yamada, #{n}rou" }
    part { "" }
  end
end
