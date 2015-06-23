class Db::ProgramsController < Db::ApplicationController
  before_action :load_work, only: [:index, :destroy]

  def index
    @programs = @work.programs.order(:started_at, :channel_id)
  end

  def destroy(id)
    @program = @work.programs.find(id)
    authorize @program, :destroy?

    @program.destroy

    redirect_to db_work_programs_path(@work), notice: "番組情報を削除しました"
  end

  private

  def load_work
    @work = Work.find(params[:work_id])
  end
end
