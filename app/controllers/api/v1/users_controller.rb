# frozen_string_literal: true

module API
  module V1
    class UsersController < API::V1::ApplicationController
      before_action :prepare_params!, only: %i(index)

      def index
        @users = API::V1::UserIndexService.new(User.all, @params).result
      end
    end
  end
end
