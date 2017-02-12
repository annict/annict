# frozen_string_literal: true

module Settings
  class AppsController < ApplicationController
    before_action :authenticate_user!

    def index
      @apps = current_user.connected_applications.available.authorized
    end

    def revoke(app_id)
      access_tokens = current_user.oauth_access_tokens.where(application_id: app_id)
      access_tokens.each(&:revoke)
      redirect_to :back, notice: t("messages.settings.apps.disconnected")
    end
  end
end
