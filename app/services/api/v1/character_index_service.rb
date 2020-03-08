# frozen_string_literal: true

module API
  module V1
    class CharacterIndexService < API::V1::BaseService
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
