# typed: false
# frozen_string_literal: true

FactoryBot.define do
  factory :forum_post_participant do
    association :forum_post
    association :user
  end
end
