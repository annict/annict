# frozen_string_literal: true

class ShareStatusToTwitterJob < ApplicationJob
  queue_as :default

  def perform(user_id, status_id)
    user = User.find(user_id)
    status = user.statuses.find(status_id)

    Deprecated::TwitterService.new(user).share!(status)
  end
end
