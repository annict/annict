# frozen_string_literal: true

module Api
  module Internal
    class ActivityGroupsController < Api::Internal::ApplicationController
      include V4::GraphqlRunnable

      before_action :authenticate_user!, only: %i(index)

      def index
        return render(nothing: true) unless user_signed_in?

        @username = params[:username]
        @page_category = params[:page_category]

        @activity_group_result = fetch_activity_group_result
      end

      private

      def fetch_activity_group_result
        if params[:page_category] == "profile-detail"
          return ProfileDetail::FetchUserActivityGroupsRepository.
            new(graphql_client: graphql_client(viewer: current_user)).
            fetch(username: params[:username], cursor: params[:cursor])
        end

        if current_user.timeline_mode.following?
          UserHome::FetchFollowingActivityGroupsRepository.
            new(graphql_client: graphql_client(viewer: current_user)).
            fetch(username: current_user.username, cursor: params[:cursor])
        else
          UserHome::FetchGlobalActivityGroupsRepository.
            new(graphql_client: graphql_client).
            fetch(cursor: params[:cursor])
        end
      end
    end
  end
end
