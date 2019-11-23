# frozen_string_literal: true

module Forum
  class HomeController < Forum::ApplicationController
    def index
      @posts = ForumPost.all.joins(:user).merge(User.without_deleted)
      @posts = localable_resources(@posts).order(last_commented_at: :desc).page(params[:page])
    end
  end
end
