# frozen_string_literal: true

module Api
  module Internal
    class FollowsController < ApplicationV6Controller
      before_action :authenticate_user!

      def create
        @user = User.only_kept.find(params[:user_id])
        current_user.follow(@user)
        EmailNotificationService.send_email("followed_user", @user, current_user.id)
        render(json: {}, status: 201)
      end

      def destroy
        @user = User.only_kept.find(params[:user_id])
        current_user.unfollow(@user)
        render(json: {}, status: 200)
      end
    end
  end
end
