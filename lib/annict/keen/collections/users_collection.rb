module Annict
  module Keen
    module Collections
      class UsersCollection < ApplicationCollection
        def create(user)
          ::Keen.delay.publish(:users, properties(:create, user))
        end

        private

        def properties(action, user)
          {
            action: action,
            user_id: user.id,
            device: browser.mobile? ? "mobile" : "pc",
            uuid: request.uuid,
            keen: { timestamp: user.created_at }
          }
        end
      end
    end
  end
end
