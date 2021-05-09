# frozen_string_literal: true

class HomeController < ApplicationController
  before_action :authenticate_user!

  def show
    set_page_category PageCategory::HOME

    @new_anime_list = Rails.cache.fetch("user-home-new-anime-list", expires_in: 3.hours) {
      Anime.order(created_at: :desc).preload(:anime_image).limit(5)
    }

    @forum_posts = Rails.cache.fetch("user-home-forum-posts", expires_in: 1.hour) {
      ForumPost.joins(:forum_category).merge(ForumCategory.with_slug(:site_news)).order(created_at: :desc).limit(5)
    }

    @activity_groups = ActivityGroup
      .preload(user: :profile)
      .joins(:user)
      .merge(current_user.followings)
      .order(created_at: :desc)
      .page(params[:page])
      .per(30)
      .without_count

    # @anime_ids = @activity_structs.flat_map { |ags| ags.items.pluck(:anime_id) }
  end
end
