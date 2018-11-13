# frozen_string_literal: true

module Forum
  class PostsController < Forum::ApplicationController
    permits :forum_category_id, :title, :body, model_name: "ForumPost"

    before_action :authenticate_user!, only: %i(new create edit update)

    def new(category: nil)
      @post = ForumPost.new
      @post.forum_category = ForumCategory.find_by(slug: category) if category.present?
    end

    def create(forum_post)
      @post = ForumPost.new(forum_post)
      @post.user = current_user
      @post.last_commented_at = Time.now
      @post.detect_locale!(:body)

      return render(:new) unless @post.valid?

      ActiveRecord::Base.transaction do
        @post.save!(validate: false)
        @post.forum_post_participants.create!(user: current_user)
        @post.notify_discord
      end

      Flash.store_data(cookies[:ann_client_uuid], notice: t("messages.forum.posts.created"))
      redirect_to forum_post_path(@post)
    end

    def show(id)
      @post = ForumPost.joins(:user).merge(User.published).find(id)
      @comments = @post.forum_comments.order(:created_at)
      @comment = @post.forum_comments.new

      store_page_params(post: @post, comments: @comments)
    end

    def edit(id)
      @post = ForumPost.find(id)
      authorize @post, :edit?
    end

    def update(id, forum_post)
      @post = ForumPost.find(id)
      authorize @post, :update?

      @post.attributes = forum_post
      @post.detect_locale!(:body)

      if @post.save
        Flash.store_data(cookies[:ann_client_uuid], notice: t("messages.forum.posts.updated"))
        redirect_to forum_post_path(@post)
      else
        render :edit
      end
    end
  end
end
