# frozen_string_literal: true

module API
  module V1
    class EpisodesController < API::V1::ApplicationController
      before_action :prepare_params!, only: [:index]

      def index
        @episodes = Episode.without_deleted.includes(:work, :prev_episode)
        @episodes = API::V1::EpisodeIndexService.new(@episodes, @params).result
      end
    end
  end
end
