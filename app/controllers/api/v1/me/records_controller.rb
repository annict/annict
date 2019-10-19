# frozen_string_literal: true

module Api
  module V1
    module Me
      class RecordsController < Api::V1::ApplicationController
        before_action :prepare_params!, only: %i(create update destroy)

        def create
          episode = Episode.published.find(@params.episode_id)
          record = episode.episode_records.new do |r|
            r.rating = @params.rating
            r.rating_state = @params.rating_state
            r.body = @params.comment
            r.shared_twitter = @params.share_twitter == "true"
            r.shared_facebook = @params.share_facebook == "true"
            r.oauth_application = doorkeeper_token.application
          end
          record.rating_state = record.rating_to_rating_state if record.rating.present?

          service = NewEpisodeRecordService.new(current_user, record)
          service.ga_client = ga_client
          service.app = doorkeeper_token.application
          service.via = "rest_api"

          begin
            service.save!
            @episode_record = service.episode_record
          rescue
            render_validation_errors service.episode_record
          end
        end

        def update
          @episode_record = current_user.episode_records.published.find(@params.id)
          @episode_record.rating = @params.rating
          @episode_record.rating_state = @params.rating_state
          @episode_record.body = @params.comment
          @episode_record.shared_twitter = @params.share_twitter == "true"
          @episode_record.shared_facebook = @params.share_facebook == "true"
          @episode_record.modify_body = true
          @episode_record.oauth_application = doorkeeper_token.application
          @episode_record.detect_locale!(:body)

          if @episode_record.valid?
            ActiveRecord::Base.transaction do
              @episode_record.save(validate: false)
              @episode_record.update_share_record_status
              @episode_record.share_to_sns
            end
          else
            render_validation_errors(@episode_record)
          end
        end

        def destroy
          @episode_record = current_user.episode_records.published.find(@params.id)
          @episode_record.record.destroy
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
