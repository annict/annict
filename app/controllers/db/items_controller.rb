class Db::ItemsController < Db::ApplicationController
  def index(work_id)
    @work = Work.find(work_id)
    @items = @work.items
  end
end
