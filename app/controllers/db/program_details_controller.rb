# frozen_string_literal: true

module Db
  class ProgramDetailsController < Db::ApplicationController
    before_action :authenticate_user!

    def index
      @work = Work.find(params[:work_id])
      @program_details = @work.program_details.order(started_at: :desc, channel_id: :asc)
    end

    def new
      @work = Work.find(params[:work_id])
      @form = DB::ProgramDetailRowsForm.new
      authorize @form, :new?
    end

    def create
      @work = Work.find(params[:work_id])
      @form = DB::ProgramDetailRowsForm.new(program_detail_rows_form_params)
      @form.user = current_user
      @form.work = @work
      authorize @form, :create?

      return render(:new) unless @form.valid?
      @form.save!

      flash[:notice] = t("messages._common.created")
      redirect_to db_work_program_details_path(@work)
    end

    def edit
      @program_detail = ProgramDetail.find(params[:id])
      authorize @program_detail, :edit?
      @work = @program_detail.work
    end

    def update
      @program_detail = ProgramDetail.find(params[:id])
      authorize @program_detail, :update?
      @work = @program_detail.work

      @program_detail.attributes = program_detail_params
      @program_detail.user = current_user

      return render(:edit) unless @program_detail.valid?
      @program_detail.save_and_create_activity!

      flash[:notice] = t("messages._common.updated")
      redirect_to db_work_program_details_path(@work)
    end

    def hide
      @program_detail = ProgramDetail.find(params[:id])
      authorize @program_detail, :hide?

      @program_detail.hide!

      flash[:notice] = t("resources.program_detail.unpublished")
      redirect_back fallback_location: db_works_path
    end

    def destroy
      @program_detail = ProgramDetail.find(params[:id])
      authorize @program_detail, :destroy?

      @program_detail.destroy

      flash[:notice] = t("resources.program_detail.deleted")
      redirect_back fallback_location: db_works_path
    end

    def activities
      @program_detail = ProgramDetail.find(params[:id])
      @activities = @program_detail.db_activities.order(id: :desc)
      @comment = @program_detail.db_comments.new
    end

    private

    def program_detail_params
      params.require(:program_detail).permit(
        :channel_id, :started_at, :time_zone, :rebroadcast, :vod_title_code, :vod_title_name,
        :minimum_episode_generatable_number
      )
    end

    def program_detail_rows_form_params
      params.require(:db_program_detail_rows_form).permit(:rows)
    end
  end
end
