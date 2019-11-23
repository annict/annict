# frozen_string_literal: true

class UsersController < ApplicationController
  before_action :load_i18n, only: %i(show following followers)
  before_action :authenticate_user!, only: %i(destroy share)
  before_action :set_user, only: %i(show following followers)

  def show
    @watching_works = @user.works.watching.without_deleted
    tracked_works = @watching_works.tracked_by(@user).order("c2.record_id DESC")
    other_works = @watching_works.where.not(id: tracked_works.pluck(:id))
    @works = (tracked_works + other_works).first(9)
    @favorite_characters = @user.favorite_characters.includes(:character).order(id: :desc)
    @favorite_casts = @user.favorite_people.with_cast.includes(:person).order(id: :desc)
    @favorite_staffs = @user.favorite_people.with_staff.includes(:person).order(id: :desc)
    @favorite_organizations = @user.favorite_organizations.includes(:organization).order(id: :desc)

    activities = UserActivitiesQuery.new.call(
      activities: @user.activities,
      user: current_user,
      page: 1
    )
    works = Work.without_deleted.where(id: activities.all.map(&:work_id))

    activity_data = render_jb("api/internal/activities/index",
      user: user_signed_in? ? current_user : nil,
      activities: activities,
      works: works)

    gon.push(activityData: activity_data)
  end

  def following
    @users = @user.followings.without_deleted.order("follows.id DESC")
  end

  def followers
    @users = @user.followers.without_deleted.order("follows.id DESC")
  end

  def destroy
    current_user.leave
    sign_out current_user
    redirect_to root_path, notice: t("messages.users.bye_bye")
  end

  private

  def set_user
    @user = User.without_deleted.find_by!(username: params[:username])
  end

  def load_i18n
    keys = {
      "verb.follow": nil,
      "noun.following": nil,
      "messages._common.are_you_sure": nil,
      "messages.components.mute_user_button.the_user_has_been_muted": nil
    }

    load_i18n_into_gon keys
  end
end
