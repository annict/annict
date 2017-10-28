# frozen_string_literal: true

module Api
  module Internal
    class RecordsController < Api::Internal::ApplicationController
      permits :episode_id, :comment, :shared_twitter, :shared_facebook, :rating_state

      before_action :authenticate_user!, only: %i(create)

      def create(record, page_category)
        episode = Episode.published.find(record[:episode_id])
        record = episode.records.new do |c|
          c.comment = record[:comment]
          c.shared_twitter = record[:shared_twitter]
          c.shared_facebook = record[:shared_facebook]
          c.rating_state = record[:rating_state]
        end
        ga_client.page_category = page_category

        service = NewRecordService.new(current_user, record)
        service.ga_client = ga_client

        begin
          service.save!
          head 201
        rescue
          @record = service.record
          render status: 400, json: { message: @record.errors.full_messages.first }
        end
      end
    end
  end
end
