# frozen_string_literal: true

module Settings
  class AppsController < ApplicationV6Controller
    before_action :authenticate_user!

    def index
      @apps = current_user.connected_applications.available.authorized
      @tokens = current_user
        .oauth_access_tokens
        .available
        .personal
        .order(created_at: :desc)
    end

    def revoke
      access_tokens = current_user.oauth_access_tokens.where(application_id: params[:app_id])
      access_tokens.each(&:revoke)

      flash[:notice] = t("messages.settings.apps.disconnected")
      redirect_back fallback_location: settings_app_list_path
    end
  end
end
