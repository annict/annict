# frozen_string_literal: true

class HomeController < ApplicationController
  include V4::GraphqlRunnable

  before_action :authenticate_user!

  def show
    set_page_category Rails.configuration.page_categories.user_home

    @forum_posts = Rails.cache.fetch("user-home-forum-posts", expires_in: 1.hour) do
      posts = ForumPost.
        joins(:forum_category).
        merge(ForumCategory.with_slug(:site_news))
      localable_resources(posts).order(created_at: :desc).limit(5)
    end

    @userland_projects = Rails.cache.fetch("user-home-userland-projects", expires_in: 12.hours) do
      UserlandProject.where(id: UserlandProject.pluck(:id).sample(3))
    end

    @activity_group_entities, @page_info_entity = if current_user.timeline_mode.following?
      UserHome::FollowingActivityGroupsRepository.new(
        graphql_client: graphql_client(viewer: current_user)
      ).execute(
        username: current_user.username,
        pagination: Annict::Pagination.new(before: params[:before], after: params[:after], per: 30)
      )
    else
      UserHome::GlobalActivityGroupsRepository.new(
        graphql_client: graphql_client
      ).execute(pagination: Annict::Pagination.new(before: params[:before], after: params[:after], per: 30))
    end
  end
end
