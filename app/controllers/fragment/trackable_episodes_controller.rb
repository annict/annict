# frozen_string_literal: true

module Fragment
  class TrackableEpisodesController < Fragment::ApplicationController
    include EpisodeRecordListSettable
    include TrackableEpisodeListSettable

    before_action :authenticate_user!

    def index
      set_trackable_episode_list
    end

    def show
      @episode = Episode.only_kept.find(params[:episode_id])
      @work = @episode.work
      @form = EpisodeRecordForm.new(episode: @episode, share_to_twitter: current_user.share_record_to_twitter?)

      set_episode_record_list(@episode)
    end
  end
end
