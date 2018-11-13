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
      @comment.detect_locale!(:body)

      unless @comment.valid?
        @comments = @post.forum_comments.order(:created_at)
        return render "forum/posts/show"
      end

      ActiveRecord::Base.transaction do
        @comment.save!(validate: false)
        @post.forum_post_participants.where(user: current_user).first_or_create!
        @post.update!(last_commented_at: Time.now)
      end

      @comment.send_notification

      Flash.store_data(cookies[:ann_client_uuid], notice: t("messages.forum.comments.created"))
      redirect_to forum_post_path(@post)
    end

    def edit
      authorize @comment, :edit?
    end

    def update(forum_comment)
      authorize @comment, :update?

      @comment.attributes = forum_comment
      @comment.detect_locale!(:body)

      if @comment.save
        Flash.store_data(cookies[:ann_client_uuid], notice: t("messages.forum.comments.updated"))
        redirect_to forum_post_path(@post)
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
