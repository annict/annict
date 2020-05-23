# frozen_string_literal: true

module Api
  module Internal
    class ActivitiesController < Api::Internal::ApplicationController
      include V4::GraphqlRunnable

      before_action :authenticate_user!, only: %i(index)

      def index
        return render(nothing: true) unless user_signed_in?

        @activity_group = UserHome::FetchActivitiesRepository.new(
          graphql_client: graphql_client(viewer: current_user)
        ).fetch(activity_group_id: params[:activity_group_id], cursor: params[:cursor])
      end
    end
  end
end
