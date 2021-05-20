# frozen_string_literal: true

module V3
  class UsersController < V3::ApplicationController
    before_action :load_i18n, only: %i[following followers]
    before_action :authenticate_user!, only: %i[destroy]
    before_action :set_user, only: %i[following followers]

    def following
      set_page_category PageCategory::FOLLOWING_LIST

      @users = @user.followings.only_kept.order("follows.id DESC")
    end

    def followers
      set_page_category PageCategory::FOLLOWER_LIST

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
end
