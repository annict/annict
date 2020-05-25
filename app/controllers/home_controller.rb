# frozen_string_literal: true

class HomeController < ApplicationController
  include V4::GraphqlRunnable

  before_action :authenticate_user!

  def show
    @forum_posts = Rails.cache.fetch("user-home-forum-posts", expires_in: 1.hour) do
      posts = ForumPost.
        joins(:forum_category).
        merge(ForumCategory.with_slug(:site_news))
      localable_resources(posts).order(created_at: :desc).limit(5)
    end

    @userland_projects = Rails.cache.fetch("user-home-userland-projects", expires_in: 12.hours) do
      UserlandProject.where(id: UserlandProject.pluck(:id).sample(3))
    end

    @activity_group_result = if current_user.timeline_mode.following?
      UserHome::FetchFollowingActivityGroupsRepository.new(
        graphql_client: graphql_client(viewer: current_user)
      ).fetch(username: current_user.username, cursor: params[:cursor])
    else
      UserHome::FetchGlobalActivityGroupsRepository.new(
        graphql_client: graphql_client
      ).fetch(cursor: params[:cursor])
    end
  end
end
