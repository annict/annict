# frozen_string_literal: true

module Api
  module V1
    class UsersController < Api::V1::ApplicationController
      before_action :prepare_params!, only: %i[index]

      def index
        @users = Deprecated::Api::V1::UserIndexService.new(User.all, @params).result
      end
    end
  end
end
