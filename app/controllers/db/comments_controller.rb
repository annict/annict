# frozen_string_literal: true

module Db
  class CommentsController < Db::ApplicationController
    permits :body, :resource_id, :resource_type, model_name: "DbComment"

    before_action :authenticate_user!
    before_action :load_db_comment, only: %i(destroy)

    def create(db_comment)
      @comment = current_user.db_comments.new(db_comment)

      return render(:new) unless @comment.valid?
      @comment.save!

      flash[:notice] = t "resources.db_comment.created"
      redirect_to "/#{@comment.resource_type.tableize}/#{@comment.resource_id}/activities"
    end

    private

    def load_db_comment
      @comment = DbComment.find(params[:id])
    end
  end
end
