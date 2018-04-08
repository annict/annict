# frozen_string_literal: true

module Api
  module V1
    module Me
      class RecordsController < Api::V1::ApplicationController
        before_action :prepare_params!, only: %i(create update destroy)

        def create
          episode = Episode.published.find(@params.episode_id)
          record = episode.records.new do |r|
            r.work_id = episode.work.id
            r.rating = @params.rating
            r.rating_state = @params.rating_state
            r.comment = @params.comment
            r.shared_twitter = @params.share_twitter == "true"
            r.shared_facebook = @params.share_facebook == "true"
            r.oauth_application = doorkeeper_token.application
          end
          record.rating_state = record.rating_to_rating_state if record.rating.present?

          service = NewRecordService.new(current_user, record)
          service.ga_client = ga_client
          service.app = doorkeeper_token.application
          service.via = "rest_api"

          begin
            service.save!
            @record = service.record
          rescue
            render_validation_errors service.record
          end
        end

        def update
          @record = current_user.records.find(@params.id)
          @record.rating = @params.rating
          @record.rating_state = @params.rating_state
          @record.comment = @params.comment
          @record.shared_twitter = @params.share_twitter == "true"
          @record.shared_facebook = @params.share_facebook == "true"
          @record.modify_comment = true
          @record.oauth_application = doorkeeper_token.application
          @record.detect_locale!(:comment)

          if @record.valid?
            ActiveRecord::Base.transaction do
              @record.save(validate: false)
              @record.update_share_record_status
              @record.share_to_sns
            end
          else
            render_validation_errors(@record)
          end
        end

        def destroy
          @record = current_user.records.find(@params.id)
          @record.destroy
          head 204
        end

        private

        def render_validation_errors(record)
          errors = record.errors.full_messages.map do |message|
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
