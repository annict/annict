# frozen_string_literal: true

module V3
  class WorksController < V3::ApplicationController
    before_action :set_cache_control_headers, only: %i(show)

    def show
      return render_404 if params[:id].to_i == 0
      @work = V3::WorkDetailQuery.new(work_id: params[:id].to_i).call
      return render_404 unless @work
    end
  end
end
