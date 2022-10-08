# frozen_string_literal: true

module Deprecated::Api
  module V1
    class OrganizationIndexService < Deprecated::Api::V1::BaseService
      def result
        @collection = filter_ids
        @collection = filter_name
        @collection = sort_id
        @collection = @collection.page(@params.page).per(@params.per_page)
        @collection
      end
    end
  end
end
