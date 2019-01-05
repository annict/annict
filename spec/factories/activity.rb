# frozen_string_literal: true

FactoryBot.define do
  factory :activity do
    association :user

    factory :create_episode_record_activity do
      recipient { create(:episode) }
      trackable { create(:episode_record, user: user, episode: recipient) }
      action { "create_episode_record" }
    end

    factory :create_status_activity do
      recipient { create(:work) }
      trackable { create(:status, user: user, work: recipient) }
      action { "create_status" }
    end
  end
end
