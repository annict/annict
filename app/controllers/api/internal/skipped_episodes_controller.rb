# frozen_string_literal: true

module Api
  module Internal
    class SkippedEpisodesController < Api::Internal::ApplicationController
      before_action :authenticate_user!, only: %i[create destroy]

      def create
        episode = Episode.only_kept.find(params[:episode_id])
        library_entry = current_user.library_entries.where(work_id: episode.work_id).first_or_create!

        library_entry.append_episode!(episode)

        head 204
      end
    end
  end
end
