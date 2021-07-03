# frozen_string_literal: true

module Api
  module V1
    module Me
      class ReviewsController < Api::V1::ApplicationController
        before_action :prepare_params!, only: %i[create update destroy]

        def create
          work = Anime.only_kept.find(@params.work_id)

          work_record_params = {
            anime: work,
            comment: @params.title.present? ? "#{@params.title}\n\n#{@params.body}" : @params.body,
            rating_animation: @params.rating_animation_state,
            rating_music: @params.rating_music_state,
            rating_story: @params.rating_story_state,
            rating_character: @params.rating_character_state,
            rating_overall: @params.rating_overall_state,
            share_to_twitter: @params.share_twitter
          }
          form = Forms::AnimeRecordForm.new(work_record_params)

          if form.invalid?
            return render_validation_error(form.errors.full_messages.first)
          end

          result = Creators::AnimeRecordCreator.new(user: current_user, form: form).call

          @work_record = current_user.anime_records.find_by!(record_id: result.record.id)
        end

        def update
          anime_record = current_user.anime_records.only_kept.find(@params.id)
          anime = anime_record.anime
          record = anime_record.record

          form = Forms::AnimeRecordForm.new(
            anime: anime,
            deprecated_title: @params.title,
            comment: @params.body,
            oauth_application: doorkeeper_token.application,
            rating_animation: @params.rating_animation_state,
            rating_character: @params.rating_character_state,
            rating_music: @params.rating_music_state,
            rating_overall: @params.rating_overall_state,
            rating_story: @params.rating_story_state,
            record: record,
            share_to_twitter: @params.share_twitter
          )

          if form.invalid?
            return render_validation_error(form.errors.full_messages.first)
          end

          result = Updaters::AnimeRecordUpdater.new(
            user: current_user,
            form: form
          ).call

          @work_record = result.record.anime_record
        end

        def destroy
          anime_record = current_user.anime_records.only_kept.find(@params.id)
          Destroyers::RecordDestroyer.new(record: anime_record.record).call
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
