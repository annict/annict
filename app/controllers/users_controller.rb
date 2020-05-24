# frozen_string_literal: true

class UsersController < ApplicationController
  include V4::GraphqlRunnable

  before_action :load_i18n, only: %i(show following followers)
  before_action :authenticate_user!, only: %i(destroy share)
  before_action :set_user, only: %i(show following followers)

  def show
    @watching_works = @user.works.watching.only_kept
    tracked_works = @watching_works.tracked_by(@user).order("c2.record_id DESC")
    other_works = @watching_works.where.not(id: tracked_works.pluck(:id))
    @works = (tracked_works + other_works).first(9)
    @character_favorites = @user.character_favorites.includes(:character).order(id: :desc)
    @cast_favorites = @user.person_favorites.with_cast.includes(:person).order(id: :desc)
    @staff_favorites = @user.person_favorites.with_staff.includes(:person).order(id: :desc)
    @organization_favorites = @user.organization_favorites.includes(:organization).order(id: :desc)

    @activity_group_result = ProfileDetail::FetchUserActivityGroupsRepository.new(
      graphql_client: graphql_client(viewer: current_user)
    ).fetch(username: current_user.username, cursor: params[:cursor])
  end

  def following
    @users = @user.followings.only_kept.order("follows.id DESC")
  end

  def followers
    @users = @user.followers.only_kept.order("follows.id DESC")
  end

  def destroy
    unless current_user.validate_to_destroy
      return redirect_back(
        fallback_location: root_path,
        alert: current_user.errors.full_messages.first
      )
    end

    ActiveRecord::Base.transaction do
      username = SecureRandom.uuid.underscore
      current_user.update_columns(username: username, email: "#{username}@example.com", deleted_at: Time.zone.now)

      current_user.providers.delete_all

      DestroyUserJob.perform_later(current_user.id)
    end

    sign_out current_user

    redirect_to root_path, notice: t("messages.users.bye_bye")
  end

  private

  def set_user
    @user = User.only_kept.find_by!(username: params[:username])
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
