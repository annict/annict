# typed: false
# frozen_string_literal: true

FactoryBot.define do
  factory :collection do
    user { association :registered_user }
    sequence(:name) { |n| "コレクション#{n}" }
    description { "コレクションの説明文です。" }
  end
end
