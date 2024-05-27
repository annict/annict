# typed: false
# frozen_string_literal: true

module Api
  module V1
    class RecordsController < Api::V1::ApplicationController
      before_action :prepare_params!, only: %i[index]

      def index
        @episode_records = EpisodeRecord.only_kept.includes(episode: :work, user: :profile).all
        @episode_records = Deprecated::Api::V1::EpisodeRecordIndexService.new(@episode_records, @params).result
      end
    end
  end
end
