# frozen_string_literal: true

module V4::Db
  class ActivitiesController < V4::Db::ApplicationController
    def index
      @activities = DbActivity
        .preload(:trackable, :root_resource, user: :profile)
        .order(id: :desc)
        .page(params[:page])
        .without_count
    end
  end
end
