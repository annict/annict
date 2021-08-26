# frozen_string_literal: true

class ShareRecordToTwitterJob < ApplicationJob
  queue_as :default

  def perform(user_id, record_id)
    user = User.only_kept.find_by(id: user_id)

    return unless user&.share_record_to_twitter?

    record = user.records.only_kept.find_by(id: record_id)

    if record
      TwitterService.new(user).share!(record)
    end
  end
end
