# frozen_string_literal: true

FactoryBot.define do
  factory :staff do
    association :resource, factory: :person
    work
    sequence(:name) { |n| "山田#{n}郎" }
    sequence(:name_en) { |n| "Yamada, #{n}rou" }
    role { :original_creator }
    role_other { "role_other_data" }
    role_other_en { "role_other_en_data" }
  end
end
