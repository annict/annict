# frozen_string_literal: true

module V3
  class WorksController < V3::ApplicationController
    before_action :set_cache_control_headers, only: %i(show)

    def show
      @work = V3::FetchWorkDetailService.new(work_id: params[:id]).call.
        to_h.dig("data", "searchWorks", "nodes").first
    end
  end
end
