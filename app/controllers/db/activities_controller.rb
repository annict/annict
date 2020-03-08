# frozen_string_literal: true

module DB
  class ActivitiesController < DB::ApplicationController
    def index
      @activities = DBActivity.
        includes(:trackable, :root_resource, user: :profile).
        order(id: :desc).
        page(params[:page])
    end
  end
end
