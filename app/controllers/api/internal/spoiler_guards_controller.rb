# frozen_string_literal: true

module Api::Internal
  class SpoilerGuardsController < Api::Internal::ApplicationController
    def show
      unless user_signed_in?
        return render json: {
          is_signed_in: false,
          episode_ids: [],
          work_ids: []
        }
      end

      tracked_anime_ids = current_user.anime_records.only_kept.pluck(:work_id).uniq
      tracked_episode_ids = current_user.episode_records.only_kept.pluck(:episode_id).uniq
      finished_anime_ids = current_user.library_entries.finished_to_watch.pluck(:work_id)
      anime_ids_in_library = current_user.library_entries.pluck(:work_id)

      render json: {
        is_signed_in: true,
        hide_record_body: current_user.hide_record_body?,
        watched_anime_ids: tracked_anime_ids + finished_anime_ids,
        anime_ids_in_library: anime_ids_in_library,
        tracked_episode_ids:  tracked_episode_ids
      }
    end
  end
end
