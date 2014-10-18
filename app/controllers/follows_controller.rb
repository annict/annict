class FollowsController < ApplicationController
  before_filter :authenticate_user!
  before_filter :set_user, only: [:create, :destroy]


  def create
    current_user.follow(@user)

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
