# frozen_string_literal: true

module Forum
  class CategoriesController < Forum::ApplicationController
    def show(id)
      @category = ForumCategory.find_by(slug: id)
    end
  end
end
