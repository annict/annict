# frozen_string_literal: true

module Api
  module V1
    module Me
      class ReviewsController < Api::V1::ApplicationController
        before_action :prepare_params!, only: %i[create update destroy]

        def create
          work = Work.only_kept.find(@params.work_id)

          form = Forms::WorkRecordForm.new(user: current_user, work: work)
          form.attributes = {
            comment: @params.title.present? ? "#{@params.title}\n\n#{@params.body}" : @params.body,
            rating_animation: @params.rating_animation_state,
            rating_music: @params.rating_music_state,
            rating_story: @params.rating_story_state,
            rating_character: @params.rating_character_state,
            rating_overall: @params.rating_overall_state,
            share_to_twitter: @params.share_twitter
          }

          if form.invalid?
            return render_validation_error(form.errors.full_messages.first)
          end

          result = Creators::WorkRecordCreator.new(user: current_user, form: form).call

          @work_record = current_user.work_records.find_by!(record_id: result.record.id)
        end

        def update
          work_record = current_user.work_records.only_kept.find(@params.id)
          work = work_record.work
          record = work_record.record

          form = Forms::WorkRecordForm.new(
            user: current_user,
            record: record,
            work: work,
            oauth_application: doorkeeper_token.application
          )
          form.attributes = {
            deprecated_title: @params.title,
            comment: @params.body,
            rating_animation: @params.rating_animation_state,
            rating_character: @params.rating_character_state,
            rating_music: @params.rating_music_state,
            rating_overall: @params.rating_overall_state,
            rating_story: @params.rating_story_state,
            share_to_twitter: @params.share_twitter
          }

          if form.invalid?
            return render_validation_error(form.errors.full_messages.first)
          end

          result = Updaters::WorkRecordUpdater.new(
            user: current_user,
            form: form
          ).call

          @work_record = result.record.work_record
        end

        def destroy
          work_record = current_user.work_records.only_kept.find(@params.id)
          Destroyers::RecordDestroyer.new(record: work_record.record).call
          head 204
        end

        private

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
