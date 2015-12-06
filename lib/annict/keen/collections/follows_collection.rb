module Annict
  module Keen
    module Collections
      class FollowsCollection < ApplicationCollection
        def create(follow)
          ::Keen.delay.publish(:follows, properties(:create, follow))
        end

        private

        def properties(action, follow)
          {
            action: action,
            user_id: follow.user_id,
            following_id: follow.following_id,
            device: browser.mobile? ? "mobile" : "pc",
            uuid: request.uuid,
            keen: { timestamp: follow.created_at }
          }
        end
      end
    end
  end
end
