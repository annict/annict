# frozen_string_literal: true

module Forum
  class PostsController < Forum::ApplicationController
    permits :forum_category_id, :title, :body, model_name: "ForumPost"

    before_action :authenticate_user!, only: %i(new create edit update)
    before_action :load_post, only: %i(show edit update)

    def new
      @post = ForumPost.new
    end

    def create(forum_post)
      @post = ForumPost.new(forum_post)
      @post.user = current_user
      @post.last_commented_at = Time.now

      return render(:new) unless @post.valid?

      ActiveRecord::Base.transaction do
        @post.save!(validate: false)
        @post.forum_post_participants.create!(user: current_user)
      end

      redirect_to forum_post_path(@post), notice: t("resources.forum_post.created")
    end

    def show
      @comments = @post.forum_comments.order(:created_at)
      @comment = @post.forum_comments.new
    end

    def edit
      authorize @post, :edit?
    end

    def update(forum_post)
      authorize @post, :update?

      if @post.update_attributes(forum_post)
        redirect_to forum_post_path(@post), notice: t("messages.forum.posts.updated")
      else
        render :edit
      end
    end

    private

    def load_post
      @post = ForumPost.find(params[:id])
    end
  end
end
