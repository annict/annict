class Db::ActivitiesController < Db::ApplicationController
  def index(page = nil)
    @activities = DbActivity.
                    includes(:trackable, user: :profile).
                    where.not(action: "edit_request_comments.create").
                    order(id: :desc).
                    page(page)
  end
end
