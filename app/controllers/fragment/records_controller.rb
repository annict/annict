# frozen_string_literal: true

module Fragment
  class RecordsController < Fragment::ApplicationController
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
  end
end
