# frozen_string_literal: true

module Api
  module V1
    class RecordsController < Api::V1::ApplicationController
      before_action :prepare_params!, only: [:index, :create]

      def index
        @records = Checkin.includes(episode: { work: :season }, user: :profile).all
        @records = Api::V1::RecordIndexService.new(@records, @params).result
      end

      def create
        episode = Episode.find(@params.episode_id)
        record = episode.checkins.new do |r|
          r.rating = @params.rating
          r.comment = @params.comment
          r.shared_twitter = @params.share_twitter == "true"
          r.shared_facebook = @params.share_facebook == "true"
        end

        service = NewRecordService.new(current_user, record, ga_client)

        if service.save
          @record = service.record
        else
          errors = service.record.errors.full_messages.map do |message|
            {
              type: "invalid_params",
              message: message,
              url: "http://example.com/docs/api/validations"
            }
          end

          render json: { errors: errors }, status: 400
        end
      end
    end
  end
end
