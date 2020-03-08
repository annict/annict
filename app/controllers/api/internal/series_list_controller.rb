# frozen_string_literal: true

module API
  module Internal
    class SeriesListController < API::Internal::ApplicationController
      def index
        @series_list = if params[:q].present?
          Series.where("name ILIKE ?", "%#{params[:q]}%").without_deleted
        else
          Series.none
        end
      end
    end
  end
end
