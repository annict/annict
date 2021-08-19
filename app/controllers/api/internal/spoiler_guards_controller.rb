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

      records = current_user.records.only_kept
      tracked_work_ids = records.only_work_record.pluck(:work_id).uniq
      tracked_episode_ids = records.only_episode_record.pluck(:episode_id).uniq
      finished_work_ids = current_user.library_entries.finished_to_watch.pluck(:work_id)
      work_ids_in_library = current_user.library_entries.pluck(:work_id)

      render json: {
        is_signed_in: true,
        hide_record_body: current_user.hide_record_body?,
        watched_work_ids: tracked_work_ids + finished_work_ids,
        work_ids_in_library: work_ids_in_library,
        tracked_episode_ids: tracked_episode_ids
      }
    end
  end
end
