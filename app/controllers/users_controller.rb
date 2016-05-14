# frozen_string_literal: true
# == Schema Information
#
# Table name: users
#
#  id                   :integer          not null, primary key
#  username             :string           not null
#  email                :string           not null
#  encrypted_password   :string           default(""), not null
#  remember_created_at  :datetime
#  sign_in_count        :integer          default(0), not null
#  current_sign_in_at   :datetime
#  last_sign_in_at      :datetime
#  current_sign_in_ip   :string
#  last_sign_in_ip      :string
#  confirmation_token   :string
#  confirmed_at         :datetime
#  confirmation_sent_at :datetime
#  created_at           :datetime
#  updated_at           :datetime
#  unconfirmed_email    :string
#  role                 :integer          not null
#  checkins_count       :integer          default(0), not null
#  notifications_count  :integer          default(0), not null
#
# Indexes
#
#  index_users_on_confirmation_token  (confirmation_token) UNIQUE
#  index_users_on_email               (email) UNIQUE
#  index_users_on_role                (role)
#  index_users_on_username            (username) UNIQUE
#

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

    render layout: "v1/application"
  end

  def works(status_kind, page: nil)
    @works = @user.works.on(status_kind).published.order_latest.page(page)

    render layout: "v1/application"
  end

  def following
    @users = @user.followings.order('follows.id DESC')

    render layout: "v1/application"
  end

  def followers
    @users = @user.followers.order('follows.id DESC')

    render layout: "v1/application"
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
