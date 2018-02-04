# frozen_string_literal: true

module Forum
  class CategoriesController < Forum::ApplicationController
    before_action :set_cache_control_headers, only: %i(show)

    def show(id, page: nil)
      @category = ForumCategory.find_by!(slug: id)
      @posts = localable_resources(@category.forum_posts).order(last_commented_at: :desc).page(page)

      set_surrogate_key_header(page_category, @category.record_key, @posts.map(&:record_key))
    end
  end
end
