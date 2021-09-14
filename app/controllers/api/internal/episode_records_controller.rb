# frozen_string_literal: true

module Api::Internal
  class EpisodeRecordsController < Api::Internal::ApplicationController
    before_action :authenticate_user!

    def create
      form = Forms::EpisodeRecordForm.new(user: current_user, episode: Episode.only_kept.find(params[:episode_id]))
      form.attributes = episode_record_form_params

      if form.invalid?
        return render(json: form.errors.full_messages, status: :unprocessable_entity)
      end

      Creators::EpisodeRecordCreator.new(user: current_user, form: form).call

      head :created
    end

    private

    def episode_record_form_params
      params.required(:forms_episode_record_form).permit(:body, :rating, :share_to_twitter, :watched_at)
    end
  end
end
