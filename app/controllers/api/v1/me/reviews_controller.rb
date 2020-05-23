# frozen_string_literal: true

module Api
  module V1
    module Me
      class ReviewsController < Api::V1::ApplicationController
        include V4::GraphqlRunnable

        before_action :prepare_params!, only: %i(create update destroy)

        def create
          work = Work.only_kept.find(@params.work_id)

          body = @params.title.present? ? "#{@params.title}\n\n#{@params.body}" : @params.body
          work_record_params = {
            body: body,
            rating_animation_state: @params.rating_animation_state,
            rating_music_state: @params.rating_music_state,
            rating_story_state: @params.rating_story_state,
            rating_character_state: @params.rating_character_state,
            rating_overall_state: @params.rating_overall_state,
            share_to_twitter: @params.share_twitter
          }

          work_record_entity, err = CreateWorkRecordRepository.new(
            graphql_client: graphql_client(viewer: current_user)
          ).create(work: work, params: work_record_params)

          if err
            return render_validation_error(err.message)
          end

          @work_record = current_user.work_records.find(work_record_entity.id)
        end

        def update
          @work_record = current_user.work_records.only_kept.find(@params.id)
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
              current_user.setting.save!
            end

            if @params.share_twitter == "true"
              current_user.share_work_record_to_twitter(@work_record)
            end
          else
            render_validation_errors(@work_record)
          end
        end

        def destroy
          @work_record = current_user.work_records.only_kept.find(@params.id)
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

        def render_validation_error(message)
          render(
            json: {
              errors: [
                {
                  type: "invalid_params",
                  message: message
                }
              ]
            },
            status: 400
          )
        end
      end
    end
  end
end
