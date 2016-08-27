# frozen_string_literal: true

module Api
  module Internal
    class RecordsController < Api::Internal::ApplicationController
      permits :episode_id, :comment, :shared_twitter, :shared_facebook, :rating

      before_action :authenticate_user!

      def create(record)
        episode = Episode.published.find(record[:episode_id])
        record = episode.checkins.new do |c|
          c.comment = record[:comment]
          c.shared_twitter = record[:shared_twitter]
          c.shared_facebook = record[:shared_facebook]
          c.rating = record[:rating]
        end
        service = NewRecordService.new(current_user, record, ga_client)

        if service.save
          head 201
        else
          @record = service.record
          render status: 400, json: { message: @record.errors.full_messages.first }
        end
      end
    end
  end
end
