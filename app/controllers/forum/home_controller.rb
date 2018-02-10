# frozen_string_literal: true

module Forum
  class HomeController < Forum::ApplicationController
    before_action :set_cache_control_headers, only: %i(index)

    def index(page: nil)
      @posts = localable_resources(ForumPost.all).order(last_commented_at: :desc).page(page)

      set_surrogate_key_header(page_category, @posts.map(&:record_key))
    end
  end
end
