class Db::EditRequestCommentsController < Db::ApplicationController
  permits :body

  before_action :set_edit_request, only: [:create]

  def create(edit_request_comment)
    @comment = @edit_request.comments.new(edit_request_comment)
    @comment.user = current_user

    if @comment.save
      flash[:notice] = "コメントを投稿しました"
      redirect_to db_edit_request_path(@edit_request)
    else
      @comments = @edit_request.comments.order(id: :asc)
      render "db/edit_requests/show"
    end
  end

  private

  def set_edit_request
    @edit_request = EditRequest.find(params[:edit_request_id]).decorate
  end
end
