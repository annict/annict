class Marie::ProgramsController < Marie::ApplicationController
  permits :channel_id, :episode_id, :started_at

  before_filter :set_work, only: [:index, :new, :create, :edit, :update, :destroy]
  before_filter :set_program, only: [:edit, :update, :destroy]


  def index
    @programs = @work.programs
  end

  def new
    @program = @work.programs.new
  end

  def create(program)
    @program = @work.programs.new(program)
    change_to_utc_datetime!

    if @program.save
      redirect_to marie_work_programs_path(@work), notice: '作成しました'
    else
      render 'new'
    end
  end

  def edit
    @program.started_at = @program.started_at.try(:+, (Time.now.utc_offset))
  end

  def update(program)
    @program.attributes = program
    change_to_utc_datetime!

    if @program.save
      redirect_to marie_work_programs_path(@work)
    else
      render 'edit'
    end
  end

  def destroy
    @program.destroy
    redirect_to marie_work_programs_path(@work), notice: '放送予定を削除しました'
  end


  private

  def set_program
    @program = @work.programs.find(params[:id])
  end

  def change_to_utc_datetime!
    @program.started_at = @program.started_at.in_time_zone('Asia/Tokyo') - 9.hours
    @program.sc_last_update = Time.now.in_time_zone('Asia/Tokyo')
  end
end
