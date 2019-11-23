# frozen_string_literal: true

module Api
  module Internal
    class FollowsController < ApplicationController
      before_action :authenticate_user!

      def create
        @user = User.without_deleted.find_by(username: params[:username])
        current_user.follow(@user)
        ga_client.page_category = params[:page_category]
        ga_client.events.create(:follows, :create, el: "User", ev: @user.id, ds: "internal_api")
        EmailNotificationService.send_email("followed_user", @user, current_user.id)
        head 201
      end

      def unfollow
        @user = User.without_deleted.find_by(username: params[:username])
        current_user.unfollow(@user)
        head 200
      end
    end
  end
end
