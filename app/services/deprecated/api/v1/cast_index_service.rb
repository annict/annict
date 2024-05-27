# typed: false
# frozen_string_literal: true

module Deprecated::Api
  module V1
    class CastIndexService < Deprecated::Api::V1::BaseService
      def result
        @collection = filter_ids
        @collection = filter_work_id
        @collection = sort_id
        @collection = sort_sort_number
        @collection = @collection.page(@params.page).per(@params.per_page)
        @collection
      end
    end
  end
end
