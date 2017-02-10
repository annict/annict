# frozen_string_literal: true

module Forum
  class HomeController < Forum::ApplicationController
    def index(page: nil)
      @posts = ForumPost.order(last_commented_at: :desc).page(page)
    end
  end
end
