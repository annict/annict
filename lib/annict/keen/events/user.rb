# frozen_string_literal: true

module Annict
  module Keen
    module Events
      class User < Annict::Keen::Events::Application
        def create(user)
          ::Keen.delay(priority: 10).publish(:users, properties(:create, user))
        end

        private

        def properties(action, user)
          {
            action: action,
            user_id: user.encoded_id,
            device: browser.device.mobile? ? "mobile" : "pc",
            request_uuid: request.uuid,
            keen: { timestamp: user.updated_at }
          }
        end
      end
    end
  end
end
