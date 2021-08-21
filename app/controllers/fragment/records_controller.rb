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

      authorize @record, :edit?

      @form = Forms::RecordForm.new(
        record: @record,
        work_id: @record.work_id,
        episode_id: @record.episode_id,
        body: @record.body,
        rating: @record.rating
      )
      @work = @form.work
      @episode = @form.episode

      @show_options = params[:show_options] == "true"
      @show_box = params[:show_box] == "true"
      @work_ids = [@work.id]
    end
  end
end
