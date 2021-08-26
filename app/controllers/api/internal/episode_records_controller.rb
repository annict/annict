# frozen_string_literal: true

module Api::Internal
  class EpisodeRecordsController < Api::Internal::ApplicationController
    before_action :authenticate_user!

    def create
      form = Forms::EpisodeRecordForm.new(episode_record_form_params)
      form.episode = Episode.only_kept.find(params[:episode_id])

      if form.invalid?
        return render(json: @form.errors.full_messages, status: :unprocessable_entity)
      end

      Creators::EpisodeRecordCreator.new(user: current_user, form: form).call

      head :created
    end

    private

    def episode_record_form_params
      params.required(:forms_episode_record_form).permit(:body, :rating, :share_to_twitter)
    end
  end
end
