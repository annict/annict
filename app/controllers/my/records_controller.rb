# frozen_string_literal: true

module My
  class RecordsController < V4::ApplicationController
    include Pundit

    before_action :authenticate_user!, only: %i(edit)

    def edit
      user = User.only_kept.find_by!(username: params[:username])
      @record = current_user.records.only_kept.find_by!(id: params[:record_id], user_id: user.id)

      authorize @record, :edit?

      if @record.episode_record?
        @form = EpisodeRecordForm.new(
          record: @record,
          episode: @record.episode_record.episode,
          comment: @record.episode_record.body,
          rating: @record.episode_record.rating_state,
          share_to_twitter: current_user.share_record_to_twitter?
        )
      else
        @form = AnimeRecordForm.new(episode: episode)
      end
    end

    def update
      user = User.only_kept.find_by!(username: params[:username])
      @record = current_user.records.only_kept.find_by!(id: params[:record_id], user_id: user.id)

      if @record.episode_record?
        @form = EpisodeRecordForm.new(episode_record_form_params)
        @form.record = @record
        @form.episode = @record.episode_record.episode

        if @form.invalid?
          return render json: @form.errors.full_messages, status: :unprocessable_entity
        end

        EpisodeRecordUpdater.new(user: current_user, form: @form).call
      else
      end

      head 204
    end

    private

    def episode_record_form_params
      params.required(:episode_record_form).permit(:comment, :rating, :share_to_twitter)
    end
  end
end
