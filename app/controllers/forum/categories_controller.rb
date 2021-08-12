# frozen_string_literal: true

module Forum
  class CategoriesController < Forum::ApplicationController
    def show
      @category = ForumCategory.find_by!(slug: params[:category_id])
      @posts = @category.forum_posts.joins(:user).merge(User.only_kept)
      @posts = localable_resources(@posts).order(last_commented_at: :desc).page(params[:page])
    end
  end
end
