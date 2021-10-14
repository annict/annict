# frozen_string_literal: true

module Fragment
  class RecordsController < Fragment::ApplicationController
    include Pundit
    include RecordListSettable

    before_action :authenticate_user!, only: %i[edit]

    def index
      set_page_category PageCategory::RECORD_LIST

      @user = User.only_kept.find_by!(username: params[:username])

      set_user_record_list(@user)
    end

    def show
      user = User.only_kept.find_by!(username: params[:username])
      @record = user.records.only_kept.find(params[:record_id])
      @work_ids = [@record.work_id]
    end

    def edit
      user = User.only_kept.find_by!(username: params[:username])
      @record = current_user.records.only_kept.find_by!(id: params[:record_id], user_id: user.id)
      @work = @record.work
      @episode = @record.episode

      authorize @record, :edit?

      @form = if @record.episode_record?
        Forms::EpisodeRecordForm.new(
          record: @record,
          episode: @record.episode_record.episode,
          comment: @record.episode_record.body,
          rating: @record.episode_record.rating_state,
          share_to_twitter: current_user.share_record_to_twitter?
        )
      else
        Forms::WorkRecordForm.new(
          record: @record,
          work: @record.work,
          comment: @record.comment,
          rating_overall: @record.work_record.rating_overall_state,
          rating_animation: @record.work_record.rating_animation_state,
          rating_character: @record.work_record.rating_character_state,
          rating_story: @record.work_record.rating_story_state,
          rating_music: @record.work_record.rating_music_state,
          share_to_twitter: current_user.share_record_to_twitter?
        )
      end

      @show_options = params[:show_options] == "true"
      @show_box = params[:show_box] == "true"
      @work_ids = [@work.id]
    end
  end
end
