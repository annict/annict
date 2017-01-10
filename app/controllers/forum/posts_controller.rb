# frozen_string_literal: true

module Forum
  class PostsController < Forum::ApplicationController
    permits :forum_category_id, :title, :body, model_name: "ForumPost"

    before_action :authenticate_user!, only: %i(new create edit update destroy)

    def new
      @post = ForumPost.new
    end

    def create(forum_post)
      @post = ForumPost.new(forum_post)
      @post.user = current_user
      @post.last_commented_at = Time.now

      return render(:new) unless @post.valid?
      @post.save!(validate: false)

      redirect_to forum_post_path(@post), notice: t("resources.forum_post.created")
    end

    def show(id)
      @post = ForumPost.find(id)
    end
  end
end
