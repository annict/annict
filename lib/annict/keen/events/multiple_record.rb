# frozen_string_literal: true

module Annict
  module Keen
    module Events
      class MultipleRecord < Annict::Keen::Events::Application
        def create(user)
          props = properties(:create, user)
          ::Keen.delay(priority: 10).publish(:multiple_records, props)
        end

        private

        def properties(action, user)
          {
            action: action,
            user_id: user.encoded_id,
            device: browser.device.mobile? ? "mobile" : "pc",
            client_uuid: @request.cookies["ann_client_uuid"],
            locale: I18n.locale,
            keen: { timestamp: user.updated_at }
          }
        end
      end
    end
  end
end
