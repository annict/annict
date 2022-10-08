# frozen_string_literal: true

module Deprecated::Api
  module V1
    class BaseService
      def initialize(collection, params)
        @collection = collection
        @params = params
      end

      private

      def filter_ids
        return @collection if @params.filter_ids.blank?
        @collection.where(id: @params.filter_ids.split(","))
      end

      def filter_work_id
        return @collection if @params.filter_work_id.blank?
        @collection.where(work_id: @params.filter_work_id)
      end

      def filter_episode_id
        return @collection if @params.filter_episode_id.blank?
        @collection.where(episode_id: @params.filter_episode_id)
      end

      def sort_id
        return @collection if @params.sort_id.blank?
        @collection.order(id: @params.sort_id)
      end

      def sort_sort_number
        return @collection if @params.sort_sort_number.blank?
        @collection.order(sort_number: @params.sort_sort_number)
      end

      def filter_season
        return @collection if @params.filter_season.blank?
        @collection.by_season(@params.filter_season)
      end

      def filter_title
        @collection if @params.filter_title.blank?
        @collection.ransack(title_or_title_kana_cont: @params.filter_title).result
      end

      def filter_name
        @collection if @params.filter_name.blank?
        @collection.ransack(name_or_name_kana_cont: @params.filter_name).result
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
