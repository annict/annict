# frozen_string_literal: true

module Api
  module Internal
    class AccessTokensController < Api::Internal::ApplicationController
      def create
        access_token = current_viewer.find_or_create_access_token_for_official_app!

        render json: {
          accessToken: access_token.token
        }
      end
    end
  end
end
