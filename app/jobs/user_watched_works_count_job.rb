# frozen_string_literal: true

class UserWatchedWorksCountJob < ApplicationJob
  queue_as :low

  def perform(user_id)
    user = User.only_kept.find(user_id)
    user.update_watched_works_count
  end
end
