# typed: false
# frozen_string_literal: true

FactoryBot.define do
  factory :userland_category do
    sequence(:name) { |n| "カテゴリ#{n}" }
    sequence(:name_en) { |n| "Category #{n}" }
    sequence(:sort_number) { |n| n }
  end
end
