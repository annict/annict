# frozen_string_literal: true

module Api::Internal
  class RecordsController < ApplicationV6Controller
    include Pundit

    before_action :authenticate_user!

    def update
      record = current_user.records.only_kept.find(params[:record_id])

      authorize record, :update?

      if record.episode_record?
        form = Forms::EpisodeRecordForm.new(user: current_user, record: record, episode: record.episode)
        form.attributes = episode_record_form_params


        Updaters::EpisodeRecordUpdater.new(user: current_user, form: form).call
      else
        form = Forms::WorkRecordForm.new(user: current_user, record: record, work: record.work)
        form.attributes = work_record_form_params

        if form.invalid?
          return render json: form.errors.full_messages, status: :unprocessable_entity
        end

        Updaters::WorkRecordUpdater.new(user: current_user, form: form).call
      end

      if form.invalid?
        return render json: form.errors.full_messages, status: :unprocessable_entity
      end

      render json: {}, status: 200
    end

    private

    def episode_record_form_params
      params.required(:forms_episode_record_form).permit(:body, :rating, :share_to_twitter, :watched_at)
    end

    def work_record_form_params
      params.required(:forms_work_record_form).permit(
        :body, :rating, :animation_rating, :character_rating, :music_rating, :story_rating, :share_to_twitter, :watched_at
      )
    end
  end
end
