# typed: false
# frozen_string_literal: true

module Deprecated::Api
  module V1
    module Me
      class SlotIndexService < Deprecated::Api::V1::BaseService
        attr_writer :user

        def result
          @collection = filter_unwatched
          @collection = filter_ids
          @collection = filter_channel_ids
          @collection = filter_work_ids
          @collection = filter_started_at_gt
          @collection = filter_started_at_lt
          @collection = filter_rebroadcast
          @collection = sort_id
          @collection = sort_started_at
          @collection = @collection.page(@params.page).per(@params.per_page)
          @collection
        end

        private

        def filter_unwatched
          unwatched = @params.filter_unwatched

          return @collection if unwatched.blank? || unwatched == "false"

          Deprecated::UserSlotsQuery.new(
            @user,
            @collection,
            watched: false
          ).call
        end

        def filter_channel_ids
          return @collection if @params.filter_channel_ids.blank?
          @collection.where(channel_id: @params.filter_channel_ids)
        end

        def filter_work_ids
          return @collection if @params.filter_work_ids.blank?
          @collection.where(work_id: @params.filter_work_ids)
        end

        def filter_started_at_gt
          return @collection if @params.filter_started_at_gt.blank?
          datetime = DateTime.parse(@params.filter_started_at_gt)
          @collection.where("started_at > ?", datetime)
        end

        def filter_started_at_lt
          return @collection if @params.filter_started_at_lt.blank?
          datetime = DateTime.parse(@params.filter_started_at_lt)
          @collection.where("started_at < ?", datetime)
        end

        def filter_rebroadcast
          return @collection if @params.filter_rebroadcast.blank?
          @collection.where(rebroadcast: (@params.filter_rebroadcast == "true"))
        end

        def filter_status
          return @collection if @params.filter_status.blank?
          @collection.merge(@user.library_entries.with_status(@params.filter_status))
        end

        def sort_started_at
          return @collection if @params.sort_started_at.blank?
          @collection.reorder(started_at: @params.sort_started_at)
        end
      end
    end
  end
end
