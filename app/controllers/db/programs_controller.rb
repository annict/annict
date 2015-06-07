class Db::ProgramsController < Db::ApplicationController
  def index(work_id)
    @work = Work.find(work_id)
    @programs = @work.programs.order(:started_at, :channel_id)
  end
end
