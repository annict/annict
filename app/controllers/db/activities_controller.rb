# frozen_string_literal: true

module Db
  class ActivitiesController < Db::ApplicationController
    def index
      @activities = DbActivity
        .preload(:trackable, :root_resource, user: :profile)
        .order(id: :desc)
        .page(params[:page])
    end
  end
end
