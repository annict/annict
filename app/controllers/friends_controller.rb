class FriendsController < ApplicationController
  before_filter :authenticate_user!

  def index(page)
    me_and_following_ids = current_user.followings.pluck(:id) << current_user.id

    friend_ids = current_user.social_friends.pluck(:id)
    not_following_friends_ids = friend_ids - me_and_following_ids
    @friends = User.where(id: not_following_friends_ids).page(page)

    user_ids = (User.pluck(:id) - me_and_following_ids).sample(20)
    @users = User.where(id: user_ids)
  end
end