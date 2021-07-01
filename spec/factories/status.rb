# frozen_string_literal: true

FactoryBot.define do
  factory :status do
    association :user
    association :anime
    kind { :watching }
  end
end
