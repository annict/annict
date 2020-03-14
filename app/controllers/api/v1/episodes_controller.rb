# frozen_string_literal: true

module Api
  module V1
    class EpisodesController < Api::V1::ApplicationController
      before_action :prepare_params!, only: [:index]

      def index
        @episodes = Episode.without_deleted.includes(:work, :prev_episode)
        @episodes = Api::V1::EpisodeIndexService.new(@episodes, @params).result
      end
    end
  end
end
