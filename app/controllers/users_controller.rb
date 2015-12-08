class UsersController < ApplicationController
  before_action :authenticate_user!, only: [:destroy, :share]
  before_action :set_user, only: [:show, :works, :following, :followers]

  def show
    @watching_works = @user.works.watching.published
    checkedin_works = @watching_works.checkedin_by(@user).order("c2.checkin_id DESC")
    other_works = @watching_works.where.not(id: checkedin_works.pluck(:id))
    @works = (checkedin_works + other_works).first(9)
    @graph_labels = Annict::Graphs::Checkins.labels
    @graph_values = Annict::Graphs::Checkins.values(@user)
  end

  def works(status_kind, page: nil)
    @works = @user.works.on(status_kind).published.order_latest.page(page)
  end

  def following
    @users = @user.followings.order('follows.id DESC')
  end

  def followers
    @users = @user.followers.order('follows.id DESC')
  end

  def destroy
    current_user.destroy
    redirect_to root_path, notice: "退会しました。(´・ω;:.."
  end

  private

  def set_user
    @user = User.find_by!(username: params[:username])
  end
end
