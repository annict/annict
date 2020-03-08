# frozen_string_literal: true

module API
  module V1
    class WorkIndexService < API::V1::BaseService
      def result
        @collection = filter_ids
        @collection = filter_season
        @collection = filter_title
        @collection = sort_id
        @collection = sort_season
        @collection = sort_watchers_count
        @collection = @collection.page(@params.page).per(@params.per_page)
        @collection
      end
    end
  end
end
