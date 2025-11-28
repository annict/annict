# typed: false
# frozen_string_literal: true

module Fragment
  class EpisodeRecordsController < Fragment::ApplicationController
    include EpisodeRecordListSettable

    before_action :authenticate_user!, only: %i[index]

    def index
      episode = Episode.only_kept.find(params[:episode_id])

      set_episode_record_list(episode)
    end
  end
end
