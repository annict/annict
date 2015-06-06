class Db::EpisodesController < Db::ApplicationController
  def index(work_id)
    @work = Work.find(work_id)
    @episodes = @work.episodes.order(:sort_number)
  end
end
