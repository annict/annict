# frozen_string_literal: true

module V3
  class WorksController < V3::ApplicationController
    def show
      @work = Work.without_deleted.find(params[:id])
    end
  end
end
