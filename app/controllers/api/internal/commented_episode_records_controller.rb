# frozen_string_literal: true

module Api::Internal
  class CommentedEpisodeRecordsController < Api::Internal::ApplicationController
    before_action :authenticate_user!, only: %i[create]

    def create
      episode = Episode.only_kept.find(params[:episode_id])
      @form = Forms::EpisodeRecordForm.new(user: current_user, episode: episode)
      @form.attributes = episode_record_form_params

      if @form.invalid?
        return render json: @form.errors.full_messages, status: :unprocessable_entity
      end

      Creators::EpisodeRecordCreator.new(user: current_user, form: @form).call

      render(json: {}, status: 201)
    end

    private

    def episode_record_form_params
      params.required(:forms_episode_record_form).permit(:comment, :rating, :share_to_twitter, :watched_at)
    end
  end
end
