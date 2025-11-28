# typed: false
# frozen_string_literal: true

module Settings
  class UsersController < ApplicationV6Controller
    before_action :authenticate_user!, only: %i[destroy]

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
  end
end
