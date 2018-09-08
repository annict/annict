# frozen_string_literal: true

module Api
  module Internal
    module V3
      class BaseDataController < Api::Internal::V3::ApplicationController
        include Analyzable
        include LocalHelper

        def show
          data = {
            viewerUUID: viewer_uuid,
            csrfParam: request_forgery_protection_token,
            csrfToken: form_authenticity_token,
            domain: local_domain,
            encodedUserId: current_user&.encoded_id.presence || "",
            env: Rails.env,
            gaTrackingId: ga_tracking_id(request),
            isSignedIn: user_signed_in?,
            locale: locale,
            userType: user_signed_in? ? "user" : "guest",
          }

          render json: data
        end
      end
    end
  end
end
