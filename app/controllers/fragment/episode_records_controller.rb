# frozen_string_literal: true

module Fragment
  class EpisodeRecordsController < Fragment::ApplicationController
    include EpisodeRecordListSettable

    before_action :authenticate_user!

    def index
      episode = Episode.only_kept.find(params[:episode_id])

      @form = EpisodeRecordForm.new(episode: episode, share_to_twitter: current_user.share_record_to_twitter?)

      set_episode_record_list(episode)
    end
  end
end
