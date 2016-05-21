# frozen_string_literal: true

FactoryGirl.define do
  factory :checkin do
    association :user
    comment "おもしろかった"
    twitter_url_hash "xxxxx"
    episode
    rating 3.0

    before(:create) do
      Tip.create_with(attributes_for(:record_tip)).
        find_or_create_by(partial_name: "checkin")
    end
    before(:create) { |c| c.work = c.episode.work }
  end
end
