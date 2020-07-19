# frozen_string_literal: true

module V4
  class UsersController < V4::ApplicationController
    def show
      set_page_category Rails.configuration.page_categories.profile_detail

      user = User.only_kept.find_by!(username: params[:username])

      @user_entity = Rails.cache.fetch(profile_user_cache_key(user), expires_in: 3.hours) do
        ProfileDetail::FetchUserRepository.new(graphql_client: graphql_client).fetch(username: user.username)
      end

      @activity_group_entities, @page_info_entity = ProfileDetail::FetchUserActivityGroupsRepository.new(
        graphql_client: graphql_client
      ).fetch(
        username: user.username,
        pagination: Annict::Pagination.new(before: params[:before], after: params[:after], per: 30)
      )
    end

    private

    def profile_user_cache_key(user)
      [
        "profile",
        "user",
        user.id,
        user.updated_at.rfc3339,
        user.records_count,
        user.watching_works_count,
        user.completed_works_count,
        user.following_count,
        user.followers_count,
        user.character_favorites_count,
        user.person_favorites_count,
        user.organization_favorites_count
      ].freeze
    end
  end
end
