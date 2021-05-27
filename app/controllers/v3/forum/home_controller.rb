# frozen_string_literal: true

module V3::Forum
  class HomeController < V3::Forum::ApplicationController
    def show
      @posts = ForumPost.all.joins(:user).merge(User.only_kept)
      @posts = localable_resources(@posts).order(last_commented_at: :desc).page(params[:page]).without_count
    end
  end
end
