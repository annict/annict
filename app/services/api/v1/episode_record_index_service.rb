# frozen_string_literal: true

module Api
  module V1
    class EpisodeRecordIndexService < Api::V1::BaseService
      def result
        @collection = filter_ids
        @collection = filter_episode_id
        @collection = filter_has_record_comment
        @collection = sort_id
        @collection = sort_likes_count
        @collection = @collection.page(@params.page).per(@params.per_page)
        @collection
      end

      private

      def filter_episode_id
        return @collection if @params.filter_episode_id.blank?
        @collection.where(records: {episode_id: @params.filter_episode_id})
      end

      def sort_likes_count
        return @collection if @params.sort_likes_count.blank?
        @collection.order("records.likes_count" => @params.sort_likes_count)
      end

      def filter_has_record_comment
        return @collection if @params.filter_has_record_comment.blank?
        return @collection.merge(Record.with_body) if @params.filter_has_record_comment == "true"
        @collection.merge(Record.with_no_body)
      end
    end
  end
end
