# frozen_string_literal: true

module Fragment
  class RecordsController < Fragment::ApplicationController
    include Pundit::Authorization
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

      if @record.episode_record?
        episode_record = @record.episode_record
        @form = Forms::EpisodeRecordForm.new(user: current_user, record: @record, episode: episode_record.episode)
        @form.attributes = {
          comment: episode_record.body,
          rating: episode_record.rating_state,
          share_to_twitter: current_user.share_record_to_twitter?,
          watched_at: @record.watched_at
        }
      else
        work_record = @record.work_record
        @form = Forms::WorkRecordForm.new(user: current_user, record: @record, work: @record.work)
        @form.attributes = {
          comment: @record.comment,
          rating_overall: work_record.rating_overall_state,
          rating_animation: work_record.rating_animation_state,
          rating_character: work_record.rating_character_state,
          rating_story: work_record.rating_story_state,
          rating_music: work_record.rating_music_state,
          share_to_twitter: current_user.share_record_to_twitter?,
          watched_at: @record.watched_at
        }
      end

      @show_options = params[:show_options] == "true"
      @show_box = params[:show_box] == "true"
      @work_ids = [@work.id]
    end
  end
end
