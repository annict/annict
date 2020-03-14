# frozen_string_literal: true

module Db
  class CommentsController < Db::ApplicationController
    before_action :authenticate_user!

    def create
      @comment = current_user.db_comments.new(db_comment_params)

      return render(:new) unless @comment.valid?
      @comment.save!

      flash[:notice] = t "messages.db.comments.created"
      path = "/db/#{@comment.resource_type.tableize}/#{@comment.resource_id}/activities"
      redirect_to path
    end

    private

    def db_comment_params
      params.require(:db_comment).permit(:body, :resource_id, :resource_type)
    end
  end
end
