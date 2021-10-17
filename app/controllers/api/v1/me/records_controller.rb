# frozen_string_literal: true

module Api
  module V1
    module Me
      class RecordsController < Api::V1::ApplicationController
        before_action :prepare_params!, only: %i[create update destroy]

        def create
          episode = Episode.only_kept.find(@params.episode_id)
          form = Forms::EpisodeRecordForm.new(user: current_user, episode: episode)
          form.attributes = {
            comment: @params.comment,
            deprecated_rating: @params.rating,
            rating: @params.rating_state,
            share_to_twitter: @params.share_twitter
          }

          if form.invalid?
            return render_validation_error(form.errors.full_messages.first)
          end

          creator = Creators::EpisodeRecordCreator.new(
            user: current_user,
            form: form
          ).call

          @episode_record = current_user.episode_records.find_by!(record_id: creator.record.id)
        end

        def update
          @episode_record = current_user.episode_records.only_kept.find(@params.id)
          episode = @episode_record.episode
          record = @episode_record.record
          oauth_application = doorkeeper_token.application

          form = Forms::EpisodeRecordForm.new(user: current_user, record: record, episode: episode, oauth_application: oauth_application)
          form.attributes = {
            comment: @params.comment,
            rating: @params.rating_state,
            deprecated_rating: @params.rating,
            share_to_twitter: @params.share_twitter
          }

          if form.invalid?
            return render_validation_error(form.errors.full_messages.first)
          end

          Updaters::EpisodeRecordUpdater.new(
            user: current_user,
            form: form
          ).call
        end

        def destroy
          @episode_record = current_user.episode_records.only_kept.find(@params.id)
          Destroyers::RecordDestroyer.new(record: @episode_record.record).call
          head 204
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
