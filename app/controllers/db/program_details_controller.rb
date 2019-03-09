# frozen_string_literal: true

module Db
  class ProgramDetailsController < Db::ApplicationController
    permits :channel_id, :started_at, :rebroadcast, :vod_title_code, :vod_title_name

    before_action :authenticate_user!
    before_action :load_work, only: %i(index new create)
    before_action :load_program_detail, only: %i(edit update hide destroy activities)

    def index
      @program_details = @work.program_details.order(id: :desc)
    end

    def new
      @form = DB::ProgramDetailRowsForm.new
      authorize @form, :new?
    end

    def create(db_program_detail_rows_form)
      @form = DB::ProgramDetailRowsForm.new(db_program_detail_rows_form.permit(:rows))
      @form.user = current_user
      @form.work = @work
      authorize @form, :create?

      return render(:new) unless @form.valid?
      @form.save!

      flash[:notice] = t("messages._common.created")
      redirect_to db_work_program_details_path(@work)
    end

    def edit
      authorize @program_detail, :edit?
      @work = @program_detail.work
    end

    def update(program_detail)
      authorize @program_detail, :update?
      @work = @program_detail.work

      @program_detail.attributes = program_detail
      @program_detail.user = current_user

      return render(:edit) unless @program_detail.valid?
      @program_detail.save_and_create_activity!

      flash[:notice] = t("messages._common.updated")
      redirect_to db_work_program_details_path(@work)
    end

    def hide
      authorize @program_detail, :hide?

      @program_detail.hide!

      flash[:notice] = t("resources.program_detail.unpublished")
      redirect_back fallback_location: db_works_path
    end

    def destroy
      authorize @program_detail, :destroy?

      @program_detail.destroy

      flash[:notice] = t("resources.program_detail.deleted")
      redirect_back fallback_location: db_works_path
    end

    def activities
      @activities = @program_detail.db_activities.order(id: :desc)
      @comment = @program_detail.db_comments.new
    end

    private

    def load_program_detail
      @program_detail = ProgramDetail.find(params[:id])
    end
  end
end
