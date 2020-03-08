# frozen_string_literal: true

module API
  module V1
    class FollowingController < API::V1::ApplicationController
      before_action :prepare_params!, only: %i(index)

      def index
        follows = Follow.includes(following: :profile)
        @follows = API::V1::FollowingIndexService.new(follows, @params).result
      end
    end
  end
end
