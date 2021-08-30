# frozen_string_literal: true

module Api
  module V1
    module Me
      class RecordsController < Api::V1::ApplicationController
        before_action :prepare_params!, only: %i[create update destroy]

        def create
          episode = Episode.only_kept.find(@params.episode_id)
          form = Forms::EpisodeRecordForm.new(
            body: @params.comment,
            advanced_rating: @params.rating,
            episode: episode,
            rating: @params.rating_state,
            share_to_twitter: @params.share_twitter
          )

          if form.invalid?
            return render_validation_error(form.errors.full_messages.first)
          end

          creator = Creators::EpisodeRecordCreator.new(
            user: current_user,
            form: form
          ).call

          @record = creator.record
          @episode_record = @record.episode_record
        end

        def update
          episode_record = EpisodeRecord.eager_load(:record).merge(current_user.records.only_kept).find(@params.id)
          record = episode_record.record
          episode = record.episode

          form = Forms::EpisodeRecordForm.new(
            advanced_rating: @params.rating,
            body: @params.comment,
            episode: episode,
            oauth_application: doorkeeper_token.application,
            rating: @params.rating_state,
            record: record,
            share_to_twitter: @params.share_twitter
          )

          if form.invalid?
            return render_validation_error(form.errors.full_messages.first)
          end

          result = Updaters::EpisodeRecordUpdater.new(
            user: current_user,
            form: form
          ).call

          @record = result.record
          @episode_record = @record.episode_record
        end

        def destroy
          episode_record = EpisodeRecord.eager_load(:record).merge(current_user.records.only_kept).find(@params.id)
          Destroyers::RecordDestroyer.new(record: episode_record.record).call
          head :no_content
        end

        private

        def render_validation_errors(record)
          errors = record.errors.full_messages.map { |message|
            {
              type: "invalid_params",
              message: message,
              url: "http://example.com/docs/api/validations"
            }
          }

          render json: {errors: errors}, status: 400
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
