# frozen_string_literal: true

class HomeController < ApplicationController
  include V4::GraphqlRunnable

  def show
    @forum_posts = ForumPost.
      joins(:forum_category).
      merge(ForumCategory.with_slug(:site_news))
    @forum_posts = localable_resources(@forum_posts).order(created_at: :desc).limit(5)

    @userland_projects = UserlandProject.where(id: UserlandProject.pluck(:id).sample(3))

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
