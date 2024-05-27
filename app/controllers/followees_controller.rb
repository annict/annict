# typed: false
# frozen_string_literal: true

class FolloweesController < ApplicationV6Controller
  def index
    set_page_category PageCategory::FOLLOWEE_LIST

    @user = User.only_kept.find_by!(username: params[:username])
    @users = @user.followings.preload(:profile).only_kept.order("follows.id DESC")
  end
end
