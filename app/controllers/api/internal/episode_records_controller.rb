# frozen_string_literal: true

module Api
  module Internal
    class EpisodeRecordsController < Api::Internal::ApplicationController
      permits :comment, :shared_twitter, :rating_state

      before_action :authenticate_user!, only: %i(create)

      def create(episode_id, episode_record, page_category)
        episode = Episode.published.find(episode_id)
        episode_record = episode.episode_records.new do |er|
          er.comment = episode_record[:comment]
          er.shared_twitter = episode_record[:shared_twitter]
          er.rating_state = episode_record[:rating_state]
        end
        ga_client.page_category = page_category
        timber.page_category = page_category

        service = NewEpisodeRecordService.new(current_user, episode_record)
        service.page_category = page_category
        service.ga_client = ga_client
        service.timber = timber
        service.via = "internal_api"

        begin
          service.save!
          head 201
        rescue => err
          render status: 400, json: { message: err.message }
        end
      end
    end
  end
end
