# frozen_string_literal: true

FactoryBot.define do
  factory :episode do
    work
    sequence(:number) { |n| "第#{n}話" }
    sequence(:title)  { |n| "Yes! プリキュア#{n}" }
  end
end
