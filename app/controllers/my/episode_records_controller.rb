# frozen_string_literal: true

module My
  class EpisodeRecordsController < V4::ApplicationController
    layout false

    before_action :authenticate_user!, only: %i(create)

    def create
      @form = EpisodeRecordForm.new(episode_record_form_params)
      @form.episode = Episode.only_kept.find(params[:episode_id])

      if @form.invalid?
        return render json: @form.errors.full_messages, status: :unprocessable_entity
      end

      EpisodeRecordCreator2.new(user: current_user, form: @form).call

      head 201
    end

    private

    def episode_record_form_params
      params.required(:episode_record_form).permit(:comment, :rating, :share_to_twitter)
    end
  end
end
