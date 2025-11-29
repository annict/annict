# typed: false
# frozen_string_literal: true

FactoryBot.define do
  factory :oauth_access_grant, class: "Oauth::AccessGrant" do
    resource_owner_id { create(:user).id }
    association :application, factory: :oauth_application
    token { SecureRandom.hex(32) }
    expires_in { 600 }
    redirect_uri { application.redirect_uri }
    scopes { "read" }
  end
end
