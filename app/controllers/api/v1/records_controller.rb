# frozen_string_literal: true

module API
  module V1
    class RecordsController < API::V1::ApplicationController
      before_action :prepare_params!, only: %i(index)

      def index
        @episode_records = EpisodeRecord.without_deleted.includes(episode: :work, user: :profile).all
        @episode_records = API::V1::EpisodeRecordIndexService.new(@episode_records, @params).result
      end
    end
  end
end
