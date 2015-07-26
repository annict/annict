class Db::EditRequestsController < Db::ApplicationController
  before_action :authenticate_user!, only: [:publish, :close]

  def index(page: nil)
    @edit_requests = EditRequest.order(id: :desc).page(page)
  end

  def show(id)
    @edit_request = EditRequest.find(id)
    @comment = @edit_request.comments.new
    @db_activities = @edit_request.db_activities
                      .where.not(action: "edit_requests.create")
                      .order(id: :asc)
  end

  def publish(id)
    @edit_request = EditRequest.find(id)
    authorize @edit_request, :publish?

    @edit_request.proposer = current_user
    @edit_request.publish!

    flash[:notice] = "編集リクエストを公開しました"
    redirect_to db_edit_request_path(@edit_request)
  end

  def close(id)
    @edit_request = EditRequest.find(id)
    authorize @edit_request, :close?

    @edit_request.proposer = current_user
    @edit_request.close!

    flash[:notice] = "編集リクエストを閉じました"
    redirect_to db_edit_request_path(@edit_request)
  end
end
