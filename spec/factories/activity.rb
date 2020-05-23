# frozen_string_literal: true

FactoryBot.define do
  factory :activity do
    association :user
    association :activity_group

    factory :create_episode_record_activity do
      itemable { create(:episode_record, user: user) }
    end

    factory :create_status_activity do
      itemable { create(:status, user: user) }
    end
  end
end
