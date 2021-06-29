# frozen_string_literal: true

FactoryBot.define do
  factory :library_entry do
    association :user
    association :anime
    association :status
    association :next_episode, factory: :episode
    watched_episode_ids { [] }
    position { 0 }
  end
end
