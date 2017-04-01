# frozen_string_literal: true

module Api
  module Internal
    class SeriesListController < Api::Internal::ApplicationController
      def index(q: nil)
        @series_list = if q.present?
          Series.where("name ILIKE ?", "%#{q}%").published
        else
          Series.none
        end
      end
    end
  end
end
