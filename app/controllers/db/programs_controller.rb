class Db::ProgramsController < Db::ApplicationController
  permits :channel_id, :episode_id, :started_at

  before_action :authenticate_user!
  before_action :load_work, only: [:index, :new, :create, :edit, :update, :destroy]

  def index
    @programs = @work.programs.order(:started_at, :channel_id)
  end

  def new
    @program = @work.programs.new
    authorize @program, :new?
  end

  def create(program)
    @program = @work.programs.new(program)
    authorize @program, :create?

    if @program.valid?
      change_to_utc_datetime!
      @program.save_and_create_db_activity(current_user, "programs.create")
      redirect_to db_work_programs_path(@work), notice: "放送予定を登録しました"
    else
      render :new
    end
  end

  def edit(id)
    @program = @work.programs.find(id)
    authorize @program, :edit?
  end

  def update(id, program)
    @program = @work.programs.find(id)
    @program.attributes = program
    authorize @program, :update?

    if @program.valid?
      change_to_utc_datetime!
      @program.save_and_create_db_activity(current_user, "programs.update")
      redirect_to db_work_programs_path(@work), notice: "放送予定を更新しました"
    else
      render :edit
    end
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

  def change_to_utc_datetime!
    @program.started_at = @program.started_at.in_time_zone("Asia/Tokyo") - 9.hours
    @program.sc_last_update = Time.now.in_time_zone("Asia/Tokyo")
  end
end
