# frozen_string_literal: true

module Oauth
  class AuthorizationsController < Doorkeeper::AuthorizationsController
    layout "v3/simple"
  end
end
