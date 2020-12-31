# frozen_string_literal: true

class HomeController < ApplicationController
  include V4::GraphqlRunnable

  before_action :authenticate_user!

  def show
    set_page_category PageCategory::USER_HOME

    @forum_posts = Rails.cache.fetch("user-home-forum-posts", expires_in: 1.hour) do
      posts = ForumPost.
        joins(:forum_category).
        merge(ForumCategory.with_slug(:site_news))
      localable_resources(posts).order(created_at: :desc).limit(5)
    end

    @userland_projects = Rails.cache.fetch("user-home-userland-projects", expires_in: 12.hours) do
      UserlandProject.where(id: UserlandProject.pluck(:id).sample(3))
    end

    result = if current_user.timeline_mode.following?
      UserHomePage::FollowingActivityGroupsRepository.new(
        graphql_client: graphql_client(viewer: current_user)
      ).execute(
        username: current_user.username,
        pagination: Annict::Pagination.new(before: params[:before], after: params[:after], per: 30)
      )
    else
      UserHomePage::GlobalActivityGroupsRepository.new(
        graphql_client: graphql_client
      ).execute(pagination: Annict::Pagination.new(before: params[:before], after: params[:after], per: 30))
    end
    @activity_group_entities = result.activity_group_entities
    @page_info_entity = result.page_info_entity
  end
end
