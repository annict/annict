# frozen_string_literal: true

module Api
  module V1
    class WorkIndexService < Api::V1::BaseService
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

      private

      def filter_season
        return @collection if @params.filter_season.blank?
        @collection.by_season(@params.filter_season)
      end

      def filter_title
        @collection if @params.filter_title.blank?
        @collection.search(title_or_title_kana_cont: @params.filter_title).result
      end

      def sort_season
        return @collection if @params.sort_season.blank?
        @collection.order_by_season(@params.sort_season)
      end

      def sort_watchers_count
        return @collection if @params.sort_watchers_count.blank?
        @collection.order(watchers_count: @params.sort_watchers_count)
      end
    end
  end
end
