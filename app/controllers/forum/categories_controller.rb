# frozen_string_literal: true

module Forum
  class CategoriesController < Forum::ApplicationController
    def show(id, page: nil)
      @category = ForumCategory.find_by!(slug: id)
      @posts = localable_resources(@category.forum_posts).order(last_commented_at: :desc).page(page)
    end
  end
end
