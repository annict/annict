# frozen_string_literal: true

FactoryBot.define do
  factory :series do
    sequence(:name) { |n| "#{n}人はプリキュア" }
    sequence(:name_ro) { |n| "#{n} ha Precure" }
    sequence(:name_en) { |n| "#{n} ha Precure" }
  end
end
