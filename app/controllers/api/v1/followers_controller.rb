# frozen_string_literal: true

module Api
  module V1
    class FollowersController < Api::V1::ApplicationController
      before_action :prepare_params!, only: %i[index]

      def index
        follows = Follow.includes(user: :profile)
        @follows = Api::V1::FollowersIndexService.new(follows, @params).result
      end
    end
  end
end
