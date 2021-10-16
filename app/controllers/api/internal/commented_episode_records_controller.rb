# frozen_string_literal: true

module Api::Internal
  class CommentedEpisodeRecordsController < Api::Internal::ApplicationController
    before_action :authenticate_user!, only: %i[create]

    def create
      @form = Forms::EpisodeRecordForm.new(episode_record_form_params)
      @form.user = current_user
      @form.episode = Episode.only_kept.find(params[:episode_id])

      if @form.invalid?
        return render json: @form.errors.full_messages, status: :unprocessable_entity
      end

      Creators::EpisodeRecordCreator.new(user: current_user, form: @form).call

      render(json: {}, status: 201)
    end

    private

    def episode_record_form_params
      params.required(:forms_episode_record_form).permit(:comment, :rating, :share_to_twitter)
    end
  end
end
