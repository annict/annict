# frozen_string_literal: true

FactoryBot.define do
  factory :activity do
    association :user
    association :activity_group

    trait :with_activity_group do
      after(:create) do |activity|
        create :activity_group, user: activity.user, itemable_type: activity.trackable_type
      end
    end

    factory :create_episode_record_activity do
      itemable { create(:record, :on_episode, user: user) }
    end

    factory :create_status_activity do
      itemable { create(:status, user: user) }
    end
  end
end
