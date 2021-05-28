# frozen_string_literal: true

module V3::Oauth
  class AuthorizationsController < Doorkeeper::AuthorizationsController
    layout "v3/simple"
  end
end
