class Db::EditRequestsController < Db::ApplicationController
  def index(page: nil)
    @edit_requests = EditRequest.order(id: :desc).page(page)
  end

  def show(id)
    @edit_request = EditRequest.find(id)
    @comment = @edit_request.comments.new
    @comments = @edit_request.comments.order(id: :asc)
  end
end
