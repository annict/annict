class Db::EditRequestCommentsController < Db::ApplicationController
  permits :body

  before_action :authenticate_user!
  before_action :load_edit_request

  def create(edit_request_comment)
    @comment = @edit_request.comments.new(edit_request_comment)
    authorize @comment, :create?
    @comment.user = current_user

    if @comment.save
      flash[:notice] = "コメントを保存しました"
      redirect_to db_edit_request_path(@edit_request)
    else
      render "db/edit_requests/show"
    end
  end

  def edit(id)
    @comment = @edit_request.comments.find(id)
    authorize @comment, :edit?
  end

  def update(id, edit_request_comment)
    @comment = @edit_request.comments.find(id)
    authorize @comment, :update?
    @comment.attributes = edit_request_comment

    if @comment.save
      flash[:notice] = "コメントを更新しました"
      redirect_to db_edit_request_path(@edit_request)
    else
      render :edit
    end
  end

  def destroy(id)
    @comment = @edit_request.comments.find(id)
    authorize @comment, :destroy?
    @comment.destroy
    flash[:notice] = "コメントを削除しました"
    redirect_to db_edit_request_path(@edit_request)
  end

  private

  def load_edit_request
    @edit_request = EditRequest.find(params[:edit_request_id])
  end
end
