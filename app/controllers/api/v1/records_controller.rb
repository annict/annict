# frozen_string_literal: true

module Api
  module V1
    class RecordsController < Api::V1::ApplicationController
      before_action :prepare_params!, only: %i[index]

      def index
        @episode_records = EpisodeRecord
          .eager_load(:record)
          .preload(record: [:work, :episode, user: :profile])
          .merge(Record.only_kept)
          .all
        @episode_records = Api::V1::EpisodeRecordIndexService.new(@episode_records, @params).result
      end
    end
  end
end
