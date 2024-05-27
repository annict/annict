# typed: false
# frozen_string_literal: true

module Db
  class ProgramsController < Db::ApplicationController
    before_action :authenticate_user!, only: %i[new create edit update destroy]

    def index
      @work = Work.without_deleted.find(params[:work_id])
      @programs = @work.programs.without_deleted.order(started_at: :desc, channel_id: :asc)
    end

    def new
      @work = Work.without_deleted.find(params[:work_id])
      @form = Deprecated::Db::ProgramRowsForm.new
      authorize @form
    end

    def create
      @work = Work.without_deleted.find(params[:work_id])
      @form = Deprecated::Db::ProgramRowsForm.new(program_rows_form_params)
      @form.user = current_user
      @form.work = @work
      authorize @form

      return render(:new, status: :unprocessable_entity) unless @form.valid?

      @form.save!

      redirect_to db_program_list_path(@work), notice: t("messages._common.created")
    end

    def edit
      @program = Program.without_deleted.find(params[:id])
      authorize @program
      @work = @program.work
    end

    def update
      @program = Program.without_deleted.find(params[:id])
      authorize @program
      @work = @program.work

      @program.attributes = program_params
      @program.user = current_user

      return render(:edit, status: :unprocessable_entity) unless @program.valid?

      @program.save_and_create_activity!

      redirect_to db_program_list_path(@work), notice: t("messages._common.updated")
    end

    def destroy
      @program = Program.without_deleted.find(params[:id])
      authorize @program

      @program.destroy_in_batches

      redirect_back(
        fallback_location: db_program_list_path(@program.work),
        notice: t("messages._common.deleted")
      )
    end

    private

    def program_params
      params.require(:program).permit(
        :channel_id, :started_at, :time_zone, :rebroadcast, :vod_title_code, :vod_title_name,
        :minimum_episode_generatable_number
      )
    end

    def program_rows_form_params
      params.require(:deprecated_db_program_rows_form).permit(:rows)
    end
  end
end
