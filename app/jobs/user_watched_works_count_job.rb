# frozen_string_literal: true

class UserWatchedWorksCountJob < ApplicationJob
  queue_as :default

  def perform(user)
    user.update_watched_works_count
  end
end
