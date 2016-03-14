# frozen_string_literal: true

module Annict
  module Keen
    module Collections
      class LikesCollection < ApplicationCollection
        def create(like)
          ::Keen.delay(priority: 10).publish(:likes, properties(:create, like))
        end

        private

        def properties(action, like)
          {
            action: action,
            user_id: like.user_id,
            recipient_id: like.recipient_id,
            recipient_type: like.recipient_type,
            device: browser.mobile? ? "mobile" : "pc",
            uuid: request.uuid,
            keen: { timestamp: like.created_at }
          }
        end
      end
    end
  end
end
