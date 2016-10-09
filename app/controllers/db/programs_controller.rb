# frozen_string_literal: true

module Db
  class ProgramsController < Db::ApplicationController
    permits :channel_id, :episode_id, :started_at, :rebroadcast, :time_zone

    before_action :authenticate_user!
    before_action :load_work, only: %i(index new create edit update destroy)

    def index
      @programs = @work.programs.order(started_at: :desc).order(:channel_id)
    end

    def new
      @form = DB::ProgramRowsForm.new
      authorize @form, :new?
    end

    def create(db_program_rows_form)
      @form = DB::ProgramRowsForm.new(db_program_rows_form.permit(:rows))
      @form.user = current_user
      @form.work = @work
      authorize @form, :create?

      return render(:new) unless @form.valid?
      @form.save!

      redirect_to db_work_programs_path(@work), notice: t("resources.program.created")
    end

    def edit(id)
      @program = @work.programs.find(id)
      authorize @program, :edit?
    end

    def update(id, program)
      @program = @work.programs.find(id)
      authorize @program, :update?

      @program.attributes = program
      @program.user = current_user

      return render(:edit) unless @program.valid?
      @program.save_and_create_activity!

      redirect_to db_work_programs_path(@work), notice: t("resources.program.updated")
    end

    def destroy(id)
      @program = @work.programs.find(id)
      authorize @program, :destroy?

      @program.destroy

      redirect_to db_work_programs_path(@work), notice: "放送予定を削除しました"
    end

    private

    def load_work
      @work = Work.find(params[:work_id])
    end
  end
end
