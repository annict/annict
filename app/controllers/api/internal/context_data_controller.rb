# frozen_string_literal: true

module Api
  module Internal
    class ContextDataController < Api::Internal::ApplicationController
      def show
        data = {
          csrfParam: request_forgery_protection_token,
          csrfToken: form_authenticity_token,
          encodedUserId: current_user&.encoded_id.presence || "",
          isSignedIn: user_signed_in?,
          userType: user_signed_in? ? "user" : "guest",
          viewerUUID: viewer_uuid
        }

        render json: data
      end
    end
  end
end
