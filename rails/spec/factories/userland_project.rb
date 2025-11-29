# typed: false
# frozen_string_literal: true

FactoryBot.define do
  factory :userland_project do
    association :userland_category
    sequence(:name) { |n| "プロジェクト#{n}" }
    sequence(:url) { |n| "https://example#{n}.com" }
    summary { "プロジェクトの概要です" }
    description { "プロジェクトの詳細な説明です" }
    available { true }

    trait :with_member do
      after(:create) do |project|
        create(:userland_project_member, userland_project: project)
      end
    end
  end
end
