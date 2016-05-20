# frozen_string_literal: true

FactoryGirl.define do
  factory :program do
    association :work
    association :episode
    started_at Time.now
    channel

    before(:create) { |p| p.work = p.episode.work }
  end
end
