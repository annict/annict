# typed: false
# frozen_string_literal: true

module Forum
  class PostsController < Forum::ApplicationController
    before_action :authenticate_user!, only: %i[new create edit update]

    def new
      @post = ForumPost.new
      @post.forum_category = ForumCategory.find_by(slug: params[:category]) if params[:category]
    end

    def create
      @post = ForumPost.new(forum_post_params)
      @post.user = current_user
      @post.last_commented_at = Time.now
      @post.detect_locale!(:body)

      return render(:new, status: :unprocessable_entity) unless @post.valid?

      ActiveRecord::Base.transaction do
        @post.save!(validate: false)
        @post.forum_post_participants.create!(user: current_user)
        @post.notify_discord
      end

      redirect_to forum_post_path(@post), notice: t("messages.forum.posts.created")
    end

    def show
      @post = ForumPost.joins(:user).merge(User.only_kept).find(params[:post_id])
      @comments = @post.forum_comments.order(:created_at)
      @comment = @post.forum_comments.new
    end

    def edit
      @post = ForumPost.find(params[:post_id])
      authorize @post, :edit?
    end

    def update
      @post = ForumPost.find(params[:post_id])
      authorize @post, :update?

      @post.attributes = forum_post_params
      @post.detect_locale!(:body)

      if @post.save
        redirect_to forum_post_path(@post), notice: t("messages.forum.posts.updated")
      else
        render :edit, status: :unprocessable_entity
      end
    end

    private

    def forum_post_params
      params.require(:forum_post).permit(:forum_category_id, :title, :body)
    end
  end
end
