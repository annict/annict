# frozen_string_literal: true

module TrackableEpisodeListSettable
  extend ActiveSupport::Concern

  def set_trackable_episode_list
    @library_entries = current_user
      .library_entries
      .with_not_deleted_work
      .watching
      .preload(:next_episode, program: :channel)
      .eager_load(:next_slot, work: :work_image)
      .merge(Work.where(no_episodes: false))
      .order("slots.started_at DESC NULLS LAST")
      .order(:position)
      .page(params[:page])
      .per(50)
  end
end
