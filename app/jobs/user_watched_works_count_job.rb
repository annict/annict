# frozen_string_literal: true

class UserWatchedWorksCountJob < ApplicationJob
  queue_as :low

  def perform(user)
    user.update_watched_works_count
  end
end
