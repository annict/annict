# frozen_string_literal: true

module Deprecated::Api
  module V1
    class WorkRecordIndexService < Deprecated::Api::V1::BaseService
      def result
        @collection = filter_ids
        @collection = filter_work_id
        @collection = filter_has_review_body
        @collection = sort_id
        @collection = sort_likes_count
        @collection = @collection.page(@params.page).per(@params.per_page)
        @collection
      end

      private

      def sort_likes_count
        return @collection if @params.sort_likes_count.blank?
        @collection.order(likes_count: @params.sort_likes_count)
      end

      def filter_has_review_body
        return @collection if @params.filter_has_review_body.blank?
        return @collection.with_body if @params.filter_has_review_body == "true"
        @collection.with_no_body
      end
    end
  end
end
