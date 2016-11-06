# frozen_string_literal: true

module Db
  class ActivitiesController < Db::ApplicationController
    def index(page = nil)
      @activities = DbActivity.
        includes(:trackable, :root_resource, user: :profile).
        order(id: :desc).
        page(page)
    end
  end
end
