# frozen_string_literal: true

FactoryBot.define do
  factory :activity do
    association :user
    association :activity_group

    factory :create_episode_record_activity do
      recipient { create(:episode) }
      trackable { create(:episode_record, user: user, episode: recipient) }
      work_id { recipient.work.id }
      episode_id { recipient.id }
      episode_record_id { trackable.id }
      action { "create_episode_record" }
    end

    factory :create_status_activity do
      recipient { create(:work) }
      trackable { create(:status, user: user, work: recipient) }
      action { "create_status" }
    end
  end
end
