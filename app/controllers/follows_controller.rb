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
#  index_follows_on_user_id_and_following_id  (user_id,following_id) UNIQUE
#

class FollowsController < ApplicationController
  before_action :authenticate_user!
  before_action :set_user, only: [:create, :destroy]


  def create
    current_user.follow(@user)
    ga_client.events.create("follows", "create")

    render status: 200, nothing: true
  end

  def destroy
    current_user.unfollow(@user)

    render status: 200, nothing: true
  end


  private

  def set_user
    @user = User.find(params[:id])
  end
end
