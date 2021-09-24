# frozen_string_literal: true

module Fragment
  class TrackableWorksController < Fragment::ApplicationController
    before_action :authenticate_user!

    def show
      @work = Work.only_kept.find(params[:work_id])
      @library_entry = current_user.library_entries.find_by!(work: @work)
      @episodes = @work
        .episodes
        .only_kept
        .where.not(id: @library_entry.watched_episode_ids)
        .order(:sort_number)
        .page(params[:page])
        .per(15)
      @programs = @work
        .programs
        .only_kept
        .eager_load(:channel)
        .merge(current_user.channels.only_kept)
        .order(:started_at)
    end
  end
end
