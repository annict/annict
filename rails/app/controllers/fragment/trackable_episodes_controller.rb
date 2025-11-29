# typed: false
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
      @form = Forms::EpisodeRecordForm.new(episode: @episode)

      set_episode_record_list(@episode)
    end
  end
end
