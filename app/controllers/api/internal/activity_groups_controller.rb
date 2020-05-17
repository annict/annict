# frozen_string_literal: true

module Api
  module Internal
    class ActivityGroupsController < Api::Internal::ApplicationController
      include V4::GraphqlRunnable

      before_action :authenticate_user!, only: %i(index)

      def index
        return render(nothing: true) unless user_signed_in?

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
  end
end
