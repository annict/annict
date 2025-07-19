# typed: false
# frozen_string_literal: true

FactoryBot.define do
  factory :email_notification do
    user
    unsubscription_key { "#{SecureRandom.uuid}-#{SecureRandom.uuid}" }
    event_followed_user { true }
    event_liked_episode_record { true }
    event_next_season_came { true }
    event_favorite_works_added { true }
    event_related_works_added { true }
  end
end