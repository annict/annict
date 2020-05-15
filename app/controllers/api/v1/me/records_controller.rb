# frozen_string_literal: true

module Api
  module V1
    module Me
      class RecordsController < Api::V1::ApplicationController
        include V4::GraphqlRunnable

        before_action :prepare_params!, only: %i(create update destroy)

        def create
          episode = Episode.only_kept.find(@params.episode_id)

          episode_record, err = CreateEpisodeRecordRepository.new(
            graphql_client: graphql_client(viewer: current_user)
          ).create(
            episode: episode,
            params: {
              rating: @params.rating,
              rating_state: @params.rating_state,
              body: @params.comment,
              share_to_twitter: @params.share_twitter
            }
          )

          if err
            return render_validation_error(err.message)
          end

          @episode_record = current_user.episode_records.find(episode_record.id)
        end

        def update
          @episode_record = current_user.episode_records.only_kept.find(@params.id)
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
          @episode_record = current_user.episode_records.only_kept.find(@params.id)
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
