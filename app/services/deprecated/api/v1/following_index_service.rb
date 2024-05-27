# typed: false
# frozen_string_literal: true

module Deprecated::Api
  module V1
    class FollowingIndexService < Deprecated::Api::V1::BaseService
      def result
        @collection = filter_user_id
        @collection = filter_username
        @collection = sort_id
        @collection = @collection.page(@params.page).per(@params.per_page)
        @collection
      end

      private

      def filter_user_id
        return @collection if @params.filter_user_id.blank?
        @collection.where(user_id: @params.filter_user_id)
      end

      def filter_username
        return @collection if @params.filter_username.blank?
        @collection
          .joins(:user)
          .where(users: {username: @params.filter_username})
      end
    end
  end
end
