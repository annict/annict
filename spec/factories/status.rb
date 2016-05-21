# frozen_string_literal: true

FactoryGirl.define do
  factory :status do
    association :user
    association :work
    kind :watching

    before(:create) do
      attrs = attributes_for(:status_tip)
      Tip.create_with(attrs).find_or_create_by(partial_name: "status")
    end
  end
end
