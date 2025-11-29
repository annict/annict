# typed: false
# frozen_string_literal: true

FactoryBot.define do
  factory :provider do
    name { "twitter" }
    uid { SecureRandom.hex }
    token { "mock_token" }
    token_secret { "mock_secret" }
  end
end
