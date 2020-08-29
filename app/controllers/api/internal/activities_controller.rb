# frozen_string_literal: true

module Api
  module Internal
    class ActivitiesController < Api::Internal::ApplicationController
      include V4::GraphqlRunnable

      def index
        @activity_group_entity = UserHome::ActivitiesRepository.new(
          graphql_client: graphql_client
        ).execute(activity_group_id: params[:activity_group_id], cursor: params[:cursor])
      end
    end
  end
end
