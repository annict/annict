# frozen_string_literal: true

module Api
  module V1
    module Me
      class ReviewsController < Api::V1::ApplicationController
        before_action :prepare_params!, only: %i(create update destroy)

        def create
          work = Work.published.find(@params.work_id)
          work_record = work.work_records.new do |r|
            r.user = current_user
            r.work = work
            r.title = @params.title
            r.body = @params.body
            r.rating_animation_state = @params.rating_animation_state
            r.rating_music_state = @params.rating_music_state
            r.rating_story_state = @params.rating_story_state
            r.rating_character_state = @params.rating_character_state
            r.rating_overall_state = @params.rating_overall_state
            r.oauth_application = doorkeeper_token.application
          end
          current_user.setting.attributes = {
            share_review_to_twitter: @params.share_twitter == "true",
            share_review_to_facebook: @params.share_facebook == "true"
          }

          service = NewWorkRecordService.new(current_user, work_record, current_user.setting)
          service.ga_client = ga_client
          service.logentries = logentries
          service.app = doorkeeper_token.application
          service.via = "rest_api"

          begin
            service.save!
            @work_record = service.work_record
          rescue
            render_validation_errors service.work_record
          end
        end

        def update
          @work_record = current_user.work_records.published.find(@params.id)
          @work_record.title = @params.title
          @work_record.body = @params.body
          @work_record.rating_animation_state = @params.rating_animation_state
          @work_record.rating_music_state = @params.rating_music_state
          @work_record.rating_story_state = @params.rating_story_state
          @work_record.rating_character_state = @params.rating_character_state
          @work_record.rating_overall_state = @params.rating_overall_state
          @work_record.modified_at = Time.now
          @work_record.oauth_application = doorkeeper_token.application
          @work_record.detect_locale!(:body)
          current_user.setting.attributes = {
            share_review_to_twitter: @params.share_twitter == "true",
            share_review_to_facebook: @params.share_facebook == "true"
          }

          if @work_record.valid?
            ActiveRecord::Base.transaction do
              @work_record.save(validate: false)
              @work_record.share_to_sns
              current_user.setting.save!
            end
          else
            render_validation_errors(@work_record)
          end
        end

        def destroy
          @work_record = current_user.work_records.published.find(@params.id)
          @work_record.record.destroy
          head 204
        end

        private

        def render_validation_errors(review)
          errors = review.errors.full_messages.map do |message|
            {
              type: "invalid_params",
              message: message
            }
          end

          render json: { errors: errors }, status: 400
        end
      end
    end
  end
end
