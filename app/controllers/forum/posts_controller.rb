# frozen_string_literal: true

module Forum
  class PostsController < Forum::ApplicationController
    permits :forum_category_id, :title, :body, model_name: "ForumPost"

    before_action :set_cache_control_headers, only: %i(show)
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
        FastlyRails.purge_by_key("forum_home_index", "forum_categories_show")
        @post.notify_discord
      end

      redirect_to forum_post_path(@post), notice: t("messages.forum.posts.created")
    end

    def show(id)
      @post = ForumPost.find(id)
      @comments = @post.forum_comments.order(:created_at)
      @comment = @post.forum_comments.new

      store_page_params(post: @post, comments: @comments)
      set_surrogate_key_header(page_category, @post.record_key, @comments.map(&:record_key))
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
        @post.purge
        redirect_to forum_post_path(@post), notice: t("messages.forum.posts.updated")
      else
        render :edit
      end
    end
  end
end
