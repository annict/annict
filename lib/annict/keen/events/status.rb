# frozen_string_literal: true

module Annict
  module Keen
    module Events
      class Status < Annict::Keen::Events::Application
        def create(user, app)
          props = properties(:create, user, app)
          ::Keen.delay(priority: 10).publish(:statuses, props)
        end

        private

        def properties(action, user, app)
          {
            action: action,
            user_id: user.encoded_id,
            device: browser.device.mobile? ? "mobile" : "pc",
            client_uuid: @request.cookies["ann_client_uuid"],
            oauth_application_uid: app&.uid,
            locale: I18n.locale,
            keen: { timestamp: user.updated_at }
          }
        end
      end
    end
  end
end
