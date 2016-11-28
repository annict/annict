# frozen_string_literal: true

module Db
  class ProgramsController < Db::ApplicationController
    permits :channel_id, :episode_id, :started_at, :rebroadcast, :time_zone

    before_action :authenticate_user!
    before_action :load_work, only: %i(index new create)
    before_action :load_program, only: %i(edit update destroy activities)

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

    def edit
      authorize @program, :edit?
      @work = @program.work
    end

    def update(program)
      authorize @program, :update?
      @work = @program.work

      @program.attributes = program
      @program.user = current_user

      return render(:edit) unless @program.valid?
      @program.save_and_create_activity!

      redirect_to db_work_programs_path(@work), notice: t("resources.program.updated")
    end

    def destroy
      authorize @program, :destroy?

      @program.destroy

      flash[:notice] = t("resources.program.deleted")
      redirect_back fallback_location: db_works_path
    end

    def activities
      @activities = @program.db_activities.order(id: :desc)
      @comment = @program.db_comments.new
    end

    private

    def load_work
      @work = Work.find(params[:work_id])
    end

    def load_program
      @program = Program.find(params[:id])
    end
  end
end
