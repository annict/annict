class FriendsController < ApplicationController
  before_filter :authenticate_user!

  def index
    me_and_following_ids = current_user.followings.pluck(:id) << current_user.id

    friend_ids = current_user.social_friends.all.pluck(:id)
    not_following_friend_ids = friend_ids - me_and_following_ids
    @friends = User.where(id: not_following_friend_ids).sample(20)

    user_ids = (User.pluck(:id) - (me_and_following_ids + @friends.map(&:id)))
    @users = User.where(id: user_ids).past_month(field: :current_sign_in_at).sample(20)
  end
end
