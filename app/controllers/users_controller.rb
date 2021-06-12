# frozen_string_literal: true

class UsersController < ApplicationController
  before_action :load_i18n, only: %i[following followers]
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

  def load_i18n
    keys = {
      "verb.follow": nil,
      "noun.following": nil,
      "messages._common.are_you_sure": nil
    }

    load_i18n_into_gon keys
  end
end
