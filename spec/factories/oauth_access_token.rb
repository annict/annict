# frozen_string_literal: true

FactoryBot.define do
  factory :oauth_access_token, class: "Doorkeeper::AccessToken" do
    association :application, factory: :oauth_application
    user { application.owner }
    sequence(:token) { |n| "token#{n}" }
    scopes "read write"
  end
end
