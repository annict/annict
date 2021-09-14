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
      @record = current_user.records.only_kept.find(params[:record_id])
      @work = @record.work
      @episode = @record.episode

      authorize @record, :edit?

      @form = Forms::EpisodeRecordForm.new(
        record: @record,
        user: current_user
      )

      if @record.episode_record?
        @form.attributes = {
          episode: @record.episode,
          body: @record.body,
          rating: @record.rating,
          share_to_twitter: current_user.share_record_to_twitter?,
          watched_at: @record.watched_at
        }
      else
        @form.attributes = {
          work: @record.work,
          body: @record.body,
          rating_overall: @record.rating,
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
