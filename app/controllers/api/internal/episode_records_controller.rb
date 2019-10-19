# frozen_string_literal: true

module Api
  module Internal
    class EpisodeRecordsController < Api::Internal::ApplicationController
      before_action :authenticate_user!, only: %i(create)

      def create
        episode = Episode.published.find(params[:episode_id])
        episode_record = episode.episode_records.new do |er|
          er.body = episode_record_params[:body]
          er.shared_twitter = episode_record_params[:shared_twitter]
          er.rating_state = episode_record_params[:rating_state]
        end
        ga_client.page_category = params[:page_category]

        service = NewEpisodeRecordService.new(current_user, episode_record)
        service.page_category = params[:page_category]
        service.ga_client = ga_client
        service.via = "internal_api"

        begin
          service.save!
          head 201
        rescue => err
          render status: 400, json: { message: err.message }
        end
      end

      private

      def episode_record_params
        params.require(:episode_record).permit(:body, :shared_twitter, :rating_state)
      end
    end
  end
end
