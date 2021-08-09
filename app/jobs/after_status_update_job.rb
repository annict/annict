# frozen_string_literal: true

class AfterStatusUpdateJob < ApplicationJob
  queue_as :low

  def perform(user_id, anime_id, prev_status_kind, new_status_kind)
    user = User.only_kept.find(user_id)
    anime = Anime.only_kept.find(anime_id)

    ActiveRecord::Base.transaction do
      user.update_watched_works_count
      user.update_works_count!(prev_status_kind, new_status_kind)
      anime.update_watchers_count!(prev_status_kind, new_status_kind)
    end
  end
end
