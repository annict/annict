# frozen_string_literal: true
# == Schema Information
#
# Table name: follows
#
#  id           :integer          not null, primary key
#  user_id      :integer          not null
#  following_id :integer          not null
#  created_at   :datetime
#  updated_at   :datetime
#
# Indexes
#
#  follows_following_id_idx          (following_id)
#  follows_user_id_following_id_key  (user_id,following_id) UNIQUE
#  follows_user_id_idx               (user_id)
#

module Api
  module Internal
    class FollowsController < ApplicationController
      before_action :authenticate_user!

      def create
        @user = User.published.find_by(username: params[:username])
        current_user.follow(@user)
        ga_client.page_category = params[:page_category]
        ga_client.events.create(:follows, :create, el: "User", ev: @user.id, ds: "internal_api")
        keen_client.publish(:follow_create, via: "internal_api")
        EmailNotificationService.send_email("followed_user", @user, current_user.id)
        head 201
      end

      def unfollow
        @user = User.published.find_by(username: params[:username])
        current_user.unfollow(@user)
        head 200
      end
    end
  end
end
