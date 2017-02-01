# frozen_string_literal: true

module Annict
  module Keen
    module Events
      class Status < Annict::Keen::Events::Application
        def create
          ::Keen.delay(priority: 10).publish(:statuses, properties(:create))
        end

        private

        def properties(action)
          {
            action: action,
            user_id: @user&.encoded_id,
            device: browser.device.mobile? ? "mobile" : "pc",
            client_uuid: @request.cookies["ann_client_uuid"],
            oauth_application_uid: @params[:app]&.uid,
            locale: I18n.locale,
            page_category: @params[:page_category],
            keen: { timestamp: @user&.updated_at }
          }
        end
      end
    end
  end
end
