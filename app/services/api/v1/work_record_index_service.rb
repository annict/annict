# frozen_string_literal: true

module Api
  module V1
    class WorkRecordIndexService < Api::V1::BaseService
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

      def filter_work_id
        return @collection if @params.filter_work_id.blank?
        @collection.where(records: {work_id: @params.filter_work_id})
      end

      def sort_likes_count
        return @collection if @params.sort_likes_count.blank?
        @collection.order("records.likes_count" => @params.sort_likes_count)
      end

      def filter_has_review_body
        return @collection if @params.filter_has_review_body.blank?
        return @collection.merge(Record.with_body) if @params.filter_has_review_body == "true"
        @collection.merge(Record.with_body)
      end
    end
  end
end
