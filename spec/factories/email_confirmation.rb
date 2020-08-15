# frozen_string_literal: true

FactoryBot.define do
  factory :email_confirmation do
    association :user
    back { "/foo/bar" }
    email { "example@example.com" }
    event { "sign_in" }
    expires_at { Time.zone.now + 2.hours }
    token { SecureRandom.uuid }
  end
end
