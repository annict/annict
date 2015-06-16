class Db::EditRequestsController < Db::ApplicationController
  def index(page: nil)
    @edit_requests = EditRequest.order(id: :desc).page(page)
  end

  def show(id)
    @edit_request = EditRequest.find(id)
    @comment = @edit_request.comments.new
    @comments = @edit_request.comments.order(id: :asc)
  end

  def publish(id)
    @edit_request = EditRequest.find(id)
    @edit_request.proposer = current_user
    @edit_request.publish!

    flash[:notice] = "編集リクエストを公開しました"
    redirect_to db_edit_request_path(@edit_request)
  end

  def close(id)
    @edit_request = EditRequest.find(id)
    @edit_request.proposer = current_user
    @edit_request.close!

    flash[:notice] = "編集リクエストを閉じました"
    redirect_to db_edit_request_path(@edit_request)
  end
end
