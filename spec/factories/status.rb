# frozen_string_literal: true

FactoryBot.define do
  factory :status do
    association :user
    association :work
    kind { :watching }
  end
end
