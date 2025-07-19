# typed: false
# frozen_string_literal: true

FactoryBot.define do
  factory :collection_item do
    user { association :registered_user }
    work { association :work }
    collection { association :collection, user: user }
    body { "コレクションアイテムの説明文です。" }
  end
end
