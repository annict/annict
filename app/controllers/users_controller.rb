# frozen_string_literal: true

class UsersController < ApplicationV6Controller
  before_action :set_user, only: %i[following followers]

  def following
    set_page_category PageCategory::FOLLOWING_LIST

    @users = @user.followings.only_kept.order("follows.id DESC")
  end

  def followers
    set_page_category PageCategory::FOLLOWER_LIST

    @users = @user.followers.only_kept.order("follows.id DESC")
  end

  private

  def set_user
    @user = User.only_kept.find_by!(username: params[:username])
  end
end
