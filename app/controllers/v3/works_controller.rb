# frozen_string_literal: true

module V3
  class WorksController < V3::ApplicationController
    def show
      @work = FetchWorkDetailService.new(work_id: params[:id]).call.
        to_h.dig("data", "searchWorks", "nodes").first
    end
  end
end
