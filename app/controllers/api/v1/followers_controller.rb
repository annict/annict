# frozen_string_literal: true

module API
  module V1
    class FollowersController < API::V1::ApplicationController
      before_action :prepare_params!, only: %i(index)

      def index
        follows = Follow.includes(user: :profile)
        @follows = API::V1::FollowersIndexService.new(follows, @params).result
      end
    end
  end
end
