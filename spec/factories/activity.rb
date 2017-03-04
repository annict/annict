# frozen_string_literal: true

FactoryGirl.define do
  factory :activity do
    association :user
    association :recipient
    association :trackable
    action "create_record"
  end
end
