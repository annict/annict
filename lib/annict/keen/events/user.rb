# frozen_string_literal: true

module Annict
  module Keen
    module Events
      class User < Annict::Keen::Events::Application
        def create
          ::Keen.delay(priority: 10).publish(:users, properties(:create))
        end

        private

        def properties(action)
          {
            action: action,
            user_id: @user&.encoded_id,
            device: browser.device.mobile? ? "mobile" : "pc",
            client_uuid: @request.cookies["ann_client_uuid"],
            keen: { timestamp: @user&.updated_at }
          }
        end
      end
    end
  end
end
