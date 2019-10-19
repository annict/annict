# frozen_string_literal: true

module Db
  class ProgramsController < Db::ApplicationController
    before_action :authenticate_user!

    def index
      @work = Work.find(params[:work_id])
      @programs = @work.programs.order(started_at: :desc, channel_id: :asc)
    end

    def new
      @work = Work.find(params[:work_id])
      @form = Db::ProgramRowsForm.new
      authorize @form, :new?
    end

    def create
      @work = Work.find(params[:work_id])
      @form = Db::ProgramRowsForm.new(program_rows_form_params)
      @form.user = current_user
      @form.work = @work
      authorize @form, :create?

      return render(:new) unless @form.valid?
      @form.save!

      flash[:notice] = t("messages._common.created")
      redirect_to db_work_programs_path(@work)
    end

    def edit
      @program = Program.find(params[:id])
      authorize @program, :edit?
      @work = @program.work
    end

    def update
      @program = Program.find(params[:id])
      authorize @program, :update?
      @work = @program.work

      @program.attributes = program_params
      @program.user = current_user

      return render(:edit) unless @program.valid?
      @program.save_and_create_activity!

      flash[:notice] = t("messages._common.updated")
      redirect_to db_work_programs_path(@work)
    end

    def hide
      @program = Program.find(params[:id])
      authorize @program, :hide?

      @program.hide!

      flash[:notice] = t("resources.program.unpublished")
      redirect_back fallback_location: db_works_path
    end

    def destroy
      @program = Program.find(params[:id])
      authorize @program, :destroy?

      @program.destroy

      flash[:notice] = t("resources.program.deleted")
      redirect_back fallback_location: db_works_path
    end

    def activities
      @program = Program.find(params[:id])
      @activities = @program.db_activities.order(id: :desc)
      @comment = @program.db_comments.new
    end

    private

    def program_params
      params.require(:program).permit(
        :channel_id, :started_at, :time_zone, :rebroadcast, :vod_title_code, :vod_title_name,
        :minimum_episode_generatable_number
      )
    end

    def program_rows_form_params
      params.require(:db_program_rows_form).permit(:rows)
    end
  end
end
