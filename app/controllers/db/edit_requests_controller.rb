class Db::EditRequestsController < Db::ApplicationController
  before_action :set_edit_request, only: [:show]

  def show
    @comment = @edit_request.comments.new
    @comments = @edit_request.comments.order(id: :asc)
  end

  private

  def set_edit_request
    @edit_request = EditRequest.find(params[:id]).decorate
  end
end
