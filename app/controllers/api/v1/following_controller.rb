# typed: false
# frozen_string_literal: true

module Api
  module V1
    class FollowingController < Api::V1::ApplicationController
      before_action :prepare_params!, only: %i[index]

      def index
        follows = Follow.includes(following: :profile)
        @follows = Deprecated::Api::V1::FollowingIndexService.new(follows, @params).result
      end
    end
  end
end
