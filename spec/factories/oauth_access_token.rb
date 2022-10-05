# frozen_string_literal: true

FactoryBot.define do
  factory :oauth_access_token, class: "Oauth::AccessToken" do
    association :application, factory: :oauth_application
    owner { application.owner }
    sequence(:token) { |n| "token#{n}" }
    scopes { "read write" }
  end
end
