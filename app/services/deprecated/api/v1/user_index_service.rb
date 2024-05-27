# typed: false
# frozen_string_literal: true

module Deprecated::Api
  module V1
    class UserIndexService < Deprecated::Api::V1::BaseService
      def result
        @collection = filter_ids
        @collection = filter_usernames
        @collection = sort_id
        @collection = @collection.page(@params.page).per(@params.per_page)
        @collection
      end

      private

      def filter_usernames
        return @collection if @params.filter_usernames.blank?
        @collection.where(username: @params.filter_usernames.split(","))
      end
    end
  end
end
