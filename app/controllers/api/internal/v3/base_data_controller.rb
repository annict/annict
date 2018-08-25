# frozen_string_literal: true

module Api
  module Internal
    module V3
      class BaseDataController < Api::Internal::V3::ApplicationController
        include LocalHelper

        def show
          data = {
            csrfParam: request_forgery_protection_token,
            csrfToken: form_authenticity_token,
            domain: local_domain,
            env: Rails.env,
            locale: locale,
            isSignedIn: user_signed_in?
          }

          render json: data
        end
      end
    end
  end
end
