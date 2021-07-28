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

      render json: {
        is_signed_in: true,
        episode_ids: current_user.episode_records.only_kept.pluck(:episode_id).uniq,
        work_ids: current_user.anime_records.only_kept.pluck(:work_id).uniq
      }
    end
  end
end
