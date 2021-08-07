# frozen_string_literal: true

FactoryBot.define do
  factory :anime_record do
    association :user, :with_profile
    anime
    record
    body { "おもしろかった" }
    locale { "ja" }
  end
end
