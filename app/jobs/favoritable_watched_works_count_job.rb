# typed: false
# frozen_string_literal: true

class FavoritableWatchedWorksCountJob < ApplicationJob
  queue_as :default

  def perform(resource, user)
    resource.update_watched_works_count(user)
  end
end
