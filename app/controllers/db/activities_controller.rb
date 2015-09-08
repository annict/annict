class Db::ActivitiesController < Db::ApplicationController
  def index(page = nil)
    @activities = DbActivity.includes(:trackable, user: :profile).order(id: :desc).page(page)
  end
end
