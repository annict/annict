module Annict
  module Keen
    module Collections
      class StatusesCollection < ApplicationCollection
        def create(status)
          ::Keen.delay.publish(:statuses, properties(:create, status))
        end

        private

        def properties(action, status)
          {
            action: action,
            user_id: status.user_id,
            work_id: status.work_id,
            kind: status.kind,
            is_first_status: status.user.statuses.initial?(status),
            device: browser.mobile? ? "mobile" : "pc",
            uuid: request.uuid,
            keen: { timestamp: status.created_at }
          }
        end
      end
    end
  end
end
