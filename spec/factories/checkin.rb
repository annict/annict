# frozen_string_literal: true

FactoryBot.define do
  factory :record do
    association :user, :with_profile
    comment "おもしろかった"
    twitter_url_hash "xxxxx"
    episode
    rating 3.0

    before(:create) do
      Tip.where(slug: "record").first_or_create(attributes_for(:record_tip))
    end
    before(:create) { |c| c.work = c.episode.work }
  end
end
