# frozen_string_literal: true

module Api
  module V1
    module Me
      class ReviewsController < Api::V1::ApplicationController
        before_action :prepare_params!, only: %i(create update destroy)

        def create
          work = Work.published.find(@params.work_id)
          review = work.reviews.new do |r|
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

          service = NewWorkRecordService.new(current_user, review, current_user.setting)
          service.ga_client = ga_client
          service.app = doorkeeper_token.application
          service.via = "rest_api"

          begin
            service.save!
            @review = service.review
          rescue
            render_validation_errors service.review
          end
        end

        def update
          @review = current_user.reviews.find(@params.id)
          @review.title = @params.title
          @review.body = @params.body
          @review.rating_animation_state = @params.rating_animation_state
          @review.rating_music_state = @params.rating_music_state
          @review.rating_story_state = @params.rating_story_state
          @review.rating_character_state = @params.rating_character_state
          @review.rating_overall_state = @params.rating_overall_state
          @review.modified_at = Time.now
          @review.oauth_application = doorkeeper_token.application
          @review.detect_locale!(:body)
          current_user.setting.attributes = {
            share_review_to_twitter: @params.share_twitter == "true",
            share_review_to_facebook: @params.share_facebook == "true"
          }

          if @review.valid?
            ActiveRecord::Base.transaction do
              @review.save(validate: false)
              @review.share_to_sns
              current_user.setting.save!
            end
          else
            render_validation_errors(@review)
          end
        end

        def destroy
          @review = current_user.reviews.find(@params.id)
          @review.destroy
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
