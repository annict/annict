# frozen_string_literal: true

module Api
  module V1
    module Me
      class RecordsController < Api::V1::ApplicationController
        before_action :prepare_params!, only: %i(create update destroy)

        def create
          episode = Episode.published.find(@params.episode_id)
          record = episode.records.new do |r|
            r.rating = @params.rating
            r.comment = @params.comment
            r.shared_twitter = @params.share_twitter == "true"
            r.shared_facebook = @params.share_facebook == "true"
            r.oauth_application = doorkeeper_token.application
          end

          service = NewRecordService.new(current_user, record, ga_client)

          if service.save
            @record = service.record
          else
            render_validation_errors(service.record)
          end
        end

        def update
          @record = current_user.records.find(@params.id)
          @record.rating = @params.rating
          @record.comment = @params.comment
          @record.shared_twitter = @params.share_twitter == "true"
          @record.shared_facebook = @params.share_facebook == "true"
          @record.modify_comment = true
          @record.oauth_application = doorkeeper_token.application

          if @record.valid?
            ActiveRecord::Base.transaction do
              @record.save(validate: false)
              @record.update_share_checkin_status
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
