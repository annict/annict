# typed: false
# frozen_string_literal: true

FactoryBot.define do
  factory :oauth_application, class: "Oauth::Application" do
    sequence(:name) { |n| "App name #{n}" }
    sequence(:uid) { |n| "uid#{n}" }
    sequence(:secret) { |n| "secret#{n}" }
    redirect_uri { "https://example.com" }
    scopes { "read write" }
    association :owner, factory: :registered_user
  end
end
