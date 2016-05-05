# frozen_string_literal: true

module Api
  module V1
    class RecordIndexService < Api::V1::BaseService
      def result
        @collection = filter_ids
        @collection = filter_episode_id
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
    end
  end
end
