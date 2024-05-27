# typed: false
# frozen_string_literal: true

class FollowersController < ApplicationV6Controller
  def index
    set_page_category PageCategory::FOLLOWER_LIST

    @user = User.only_kept.find_by!(username: params[:username])
    @users = @user.followers.preload(:profile).only_kept.order("follows.id DESC")
  end
end
