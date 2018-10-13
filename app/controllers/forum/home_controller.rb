# frozen_string_literal: true

module Forum
  class HomeController < Forum::ApplicationController
    def index(page: nil)
      @posts = ForumPost.all.joins(:user).merge(User.published)
      @posts = localable_resources(@posts).order(last_commented_at: :desc).page(page)
    end
  end
end
