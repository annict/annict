# typed: false
# frozen_string_literal: true

FactoryBot.define do
  factory :activity_group do
    association :user
    itemable_type { "Status" }
    single { false }
  end
end
