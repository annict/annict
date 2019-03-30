# frozen_string_literal: true

module Forum
  class CategoriesController < Forum::ApplicationController
    def show
      @category = ForumCategory.find_by!(slug: params[:id])
      @posts = localable_resources(@category.forum_posts).order(last_commented_at: :desc).page(params[:page])
    end
  end
end
