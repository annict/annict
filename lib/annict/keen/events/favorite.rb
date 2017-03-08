# frozen_string_literal: true

module Annict
  module Keen
    module Events
      class Favorite < Annict::Keen::Events::Application
        def create(attrs)
          ::Keen.delay(priority: 10).publish(:favorites, properties(:create, attrs))
        end

        private

        def properties(action, attrs)
          {
            action: action,
            user_id: @user&.encoded_id,
            device: browser.device.mobile? ? "mobile" : "pc",
            client_uuid: @request.cookies["ann_client_uuid"],
            locale: @user&.locale,
            time_zone: @user&.time_zone,
            page_category: @params[:page_category],
            resource_type: attrs[:resource_type],
            keen: { timestamp: @user&.updated_at }
          }
        end
      end
    end
  end
end
