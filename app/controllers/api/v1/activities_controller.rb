# frozen_string_literal: true

module API
  module V1
    class ActivitiesController < API::V1::ApplicationController
      before_action :prepare_params!, only: %i(index)

      def index
        activities = Activity.all.includes(:recipient, :trackable, user: :profile)
        @activities = API::V1::ActivityIndexService.new(activities, @params).result
      end
    end
  end
end
