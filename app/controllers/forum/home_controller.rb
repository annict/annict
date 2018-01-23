# frozen_string_literal: true

module Forum
  class HomeController < Forum::ApplicationController
    def index(page: nil)
      @posts = localable_resources(ForumPost.all).order(last_commented_at: :desc).page(page)
    end
  end
end
