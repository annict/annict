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
            deprecated_title: @params.title,
            body: @params.body,
            rating: @params.rating_overall_state,
            animation_rating: @params.rating_animation_state,
            music_rating: @params.rating_music_state,
            story_rating: @params.rating_story_state,
            character_rating: @params.rating_character_state,
            share_to_twitter: @params.share_twitter
          }

          if form.invalid?
            return render_validation_error(form.errors.full_messages.first)
          end

          creator = Creators::WorkRecordCreator.new(user: current_user, form: form).call

          @record = creator.record
          @work_record = @record.work_record
        end

        def update
          work_record = WorkRecord.eager_load(:record).merge(current_user.records.only_kept).find(@params.id)
          record = work_record.record
          work = record.work

          form = Forms::WorkRecordForm.new(user: current_user, work: work, record: record, oauth_application: doorkeeper_token.application)
          form.attributes = {
            deprecated_title: @params.title,
            body: @params.body,
            rating: @params.rating_overall_state,
            animation_rating: @params.rating_animation_state,
            music_rating: @params.rating_music_state,
            story_rating: @params.rating_story_state,
            character_rating: @params.rating_character_state,
            share_to_twitter: @params.share_twitter,
            watched_at: record.watched_at
          }

          if form.invalid?
            return render_validation_error(form.errors.full_messages.first)
          end

          result = Updaters::WorkRecordUpdater.new(
            user: current_user,
            form: form
          ).call

          @record = result.record
          @work_record = @record.work_record
        end

        def destroy
          work_record = WorkRecord.eager_load(:record).merge(current_user.records.only_kept).find(@params.id)
          Destroyers::RecordDestroyer.new(record: work_record.record).call
          head :no_content
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
