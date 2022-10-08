# frozen_string_literal: true

module Api
  module V1
    class ActivitiesController < Api::V1::ApplicationController
      before_action :prepare_params!, only: %i[index]

      def index
        activities = Activity.all.includes(:itemable, user: :profile)
        @activities = Deprecated::Api::V1::ActivityIndexService.new(activities, @params).result
      end
    end
  end
end
