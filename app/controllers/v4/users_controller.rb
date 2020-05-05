# frozen_string_literal: true

module V4
  class UsersController < V4::ApplicationController
    def show
      user = User.only_kept.find_by!(username: params[:username])

      @user = Rails.cache.fetch(profile_user_cache_key(user), expires_in: 3.hours) do
        UserDetail::UserRepository.new(graphql_client: graphql_client).fetch(username: user.username)
      end
    end

    private

    def profile_user_cache_key(user)
      [
        "profile",
        "user",
        user.id,
        user.updated_at.rfc3339
      ].freeze
    end
  end
end
