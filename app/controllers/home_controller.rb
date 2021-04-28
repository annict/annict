# frozen_string_literal: true

class HomeController < ApplicationController
  before_action :authenticate_user!

  def show
    set_page_category PageCategory::USER_HOME

    @new_anime_list = Rails.cache.fetch("user-home-new-anime-list", expires_in: 3.hours) do
      Anime.order(created_at: :desc).limit(4)
    end

    @forum_posts = Rails.cache.fetch("user-home-forum-posts", expires_in: 1.hour) do
      ForumPost.joins(:forum_category).merge(ForumCategory.with_slug(:site_news)).order(created_at: :desc).limit(5)
    end

    @activity_groups = if current_user.timeline_mode.following?
      ActivityGroup.eager_load(user: :profile).preload(:activities).merge(current_user.followings).order(created_at: :desc).page(params[:page]).per(30)
    else
      UserHomePage::GlobalActivityGroupsRepository.new(
        graphql_client: graphql_client
      ).execute(pagination: Annict::Pagination.new(before: params[:before], after: params[:after], per: 30))
    end
    @activity_group_structs = Builder::Activity.build(@activity_groups)
  end
end
