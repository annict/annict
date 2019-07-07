# frozen_string_literal: true

module V3
  class WorksController < V3::ApplicationController
    before_action :set_cache_control_headers, only: %i(show)

    def show
      @work = WorkDetailQuery.new(work_id: params[:id]).call
    end
  end
end
