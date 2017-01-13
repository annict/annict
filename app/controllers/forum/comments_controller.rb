# frozen_string_literal: true

module Forum
  class CommentsController < Forum::ApplicationController
    permits :body, model_name: "ForumComment"

    before_action :authenticate_user!, only: %i(create edit update)
    before_action :load_post, only: %i(create edit update)
    before_action :load_comment, only: %i(edit update)

    def create(forum_comment)
      @comment = @post.forum_comments.new(forum_comment)
      @comment.user = current_user

      if @comment.save
        redirect_to forum_post_path(@post), notice: t("messages.forum.comments.created")
      else
        @comments = @post.forum_comments.order(:created_at)
        render "forum/posts/show"
      end
    end

    def edit
      authorize @comment, :edit?
    end

    def update(forum_comment)
      authorize @comment, :update?

      if @comment.update_attributes(forum_comment)
        redirect_to forum_post_path(@post), notice: t("messages.forum.comments.updated")
      else
        render :edit
      end
    end

    private

    def load_post
      @post = ForumPost.find(params[:post_id])
    end

    def load_comment
      @comment = @post.forum_comments.find(params[:id])
    end
  end
end
