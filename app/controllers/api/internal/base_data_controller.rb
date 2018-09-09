# frozen_string_literal: true

module Api
  module Internal
    class BaseDataController < Api::Internal::ApplicationController
      include Analyzable
      include LocalHelper

      def show
        data = {
          csrfParam: request_forgery_protection_token,
          csrfToken: form_authenticity_token,
          domain: local_domain,
          encodedUserId: current_user&.encoded_id.presence || "",
          env: Rails.env,
          gaTrackingId: ga_tracking_id(request),
          isSignedIn: user_signed_in?,
          locale: locale,
          userType: user_signed_in? ? "user" : "guest",
          viewerUUID: viewer_uuid
        }

        render json: data
      end
    end
  end
end
