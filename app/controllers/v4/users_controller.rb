# frozen_string_literal: true

module V4
  class UsersController < ApplicationController
    def show
      set_page_category PageCategory::PROFILE

      user = User.only_kept.find_by!(username: params[:username])

      result = Rails.cache.fetch(profile_user_cache_key(user), expires_in: 3.hours) {
        ProfilePage::UserRepository.new(graphql_client: graphql_client).execute(username: user.username)
      }
      @user_entity = result.user_entity

      result = ProfilePage::UserActivityGroupsRepository.new(
        graphql_client: graphql_client
      ).execute(
        username: user.username,
        pagination: Annict::Pagination.new(before: params[:before], after: params[:after], per: 30)
      )
      @activity_group_entities = result.activity_group_entities
      @page_info_entity = result.page_info_entity
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
