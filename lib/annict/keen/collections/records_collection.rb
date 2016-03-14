# frozen_string_literal: true

module Annict
  module Keen
    module Collections
      class RecordsCollection < ApplicationCollection
        def create(record)
          ::Keen.delay(priority: 10).publish(:records, properties(:create, record))
        end

        private

        def properties(action, record)
          {
            action: action,
            user_id: record.user_id,
            work_id: record.episode.work_id,
            episode_id: record.episode_id,
            has_comment: record.comment.present?,
            shared_sns: record.shared_sns?,
            is_first_record: record.user.checkins.initial?(record),
            device: browser.mobile? ? "mobile" : "pc",
            uuid: request.uuid,
            keen: { timestamp: record.created_at }
          }
        end
      end
    end
  end
end
