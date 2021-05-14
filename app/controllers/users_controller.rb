# frozen_string_literal: true

class UsersController < ApplicationController
  before_action :load_i18n, only: %i[following followers]
  before_action :authenticate_user!, only: %i[destroy share]
  before_action :set_user, only: %i[following followers]

  def show
    set_page_category PageCategory::PROFILE

    @user = User.only_kept.find_by!(username: params[:username])
    @profile = @user.profile

    @activity_groups = @user
      .activity_groups
      .order(created_at: :desc)
      .page(params[:page])
      .per(30)
      .without_count

    @anime_ids = if @activity_groups.present?
      @activity_groups.flat_map.with_prelude { |ags| ags.first_item.anime_id }.uniq
    else
      []
    end
  end

  def following
    set_page_category PageCategory::FOLLOWING_LIST

    @user_entity = UserEntity.from_model(@user)
    @users = @user.followings.only_kept.order("follows.id DESC")
  end

  def followers
    set_page_category PageCategory::FOLLOWER_LIST

    @user_entity = UserEntity.from_model(@user)
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

  def profile_user_cache_key(user)
    [
      "profile",
      "user",
      user.id,
      user.updated_at.rfc3339,
      user.records_count,
      user.watching_works_count,
      user.completed_works_count,
      user.following_count,
      user.followers_count,
      user.character_favorites_count,
      user.person_favorites_count,
      user.organization_favorites_count
    ].freeze
  end
end
