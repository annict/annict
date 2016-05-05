# frozen_string_literal: true

module Api
  module V1
    class BaseService
      def initialize(collection, params)
        @collection = collection
        @params = params
      end

      private

      def filter_ids
        return @collection if @params.filter_ids.blank?
        @collection.where(id: @params.filter_ids)
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
    end
  end
end
