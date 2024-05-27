# typed: false
# frozen_string_literal: true

module Deprecated::Api
  module V1
    module Me
      class WorkIndexService < Deprecated::Api::V1::BaseService
        attr_writer :user

        def result
          @collection = filter_ids
          @collection = filter_season
          @collection = filter_title
          @collection = filter_status
          @collection = sort_id
          @collection = sort_season
          @collection = sort_watchers_count
          @collection = @collection.page(@params.page).per(@params.per_page)
          @collection
        end

        private

        def filter_status
          return @collection if @params.filter_status.blank?
          @collection.merge(@user.library_entries.with_status(@params.filter_status))
        end
      end
    end
  end
end
