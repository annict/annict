class Db::EditRequestCommentsController < Db::ApplicationController
  permits :body

  before_action :authenticate_user!

  def create(edit_request_comment)
    @edit_request = EditRequest.find(params[:edit_request_id])
    @comment = @edit_request.comments.new(edit_request_comment)
    @comment.user = current_user

    if @comment.save
      flash[:notice] = "コメントを保存しました"
      redirect_to db_edit_request_path(@edit_request)
    else
      render "db/edit_requests/show"
    end
  end
end
