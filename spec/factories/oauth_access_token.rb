# frozen_string_literal: true

FactoryGirl.define do
  factory :oauth_access_token, class: "Doorkeeper::AccessToken" do
    association :application, factory: :oauth_application
    owner { application.owner }
    sequence(:token) { |n| "token#{n}" }
    scopes "read write"
  end
end
