# frozen_string_literal: true

FactoryBot.define do
  factory :status do
    association :user
    association :work
    kind { :watching }

    before(:create) do
      attrs = attributes_for(:status_tip)
      Tip.create(attrs)
    end
  end
end
