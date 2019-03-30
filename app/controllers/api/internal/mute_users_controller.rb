# frozen_string_literal: true

module Api
  module Internal
    class MuteUsersController < Api::Internal::ApplicationController
      before_action :authenticate_user!

      def create
        user = User.find(params[:user_id])

        if current_user.mute(user)
          head :created
        else
          head :bad_request
        end
      end
    end
  end
end
