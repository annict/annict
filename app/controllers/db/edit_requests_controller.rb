class Db::EditRequestsController < Db::ApplicationController
  def show(id)
    @edit_request = EditRequest.find(id)
    @comment = @edit_request.comments.new
    @comments = @edit_request.comments.order(id: :asc)
  end
end
