# frozen_string_literal: true

class HomeController < ApplicationController
  before_action :authenticate_user!

  def show
    set_page_category PageCategory::HOME

    @new_anime_list = Rails.cache.fetch("user-home-new-anime-list", expires_in: 3.hours) {
      Anime.order(created_at: :desc).preload(:anime_image).limit(6)
    }

    @forum_posts = Rails.cache.fetch("user-home-forum-posts", expires_in: 1.hour) {
      ForumPost.joins(:forum_category).merge(ForumCategory.with_slug(:site_news)).order(created_at: :desc).limit(5)
    }

    # activity_group_mapping = {
    #   id: "activity_groups.id",
    #   itemable_type: "activity_groups.itemable_type",
    #   activities_count: "activity_groups.activities_count"
    # }
    @activity_groups = if current_user.timeline_mode.following?
      ActivityGroup
        .preload(user: :profile)
        .joins(:user)
        .merge(current_user.followings)
        .order(created_at: :desc)
        .page(params[:page])
        .per(30)
    else
      UserHomePage::GlobalActivityGroupsRepository.new(
        graphql_client: graphql_client
      ).execute(pagination: Annict::Pagination.new(before: params[:before], after: params[:after], per: 30))
    end

    # @activity_group_structs = @activity_groups.map do |ag|
    #   binding.pry
    #   Builder::Activity::ActivityGroupStruct.new(activity_group_mapping.keys.zip(ag).to_h)
    # end
    # @activity_structs = ActivityStructs::Builder.new(activity_groups: @activity_groups, current_user: current_user).call.compact

    # @anime_ids = @activity_structs.flat_map { |ags| ags.items.pluck(:anime_id) }
    # @anime_ids += @new_anime_list.pluck(:id)
  end
end
