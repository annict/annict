# frozen_string_literal: true

class AppsController < ApplicationController
  before_action :authenticate_user!

  def index
    @access_tokens = current_user.oauth_access_tokens.where(revoked_at: nil)
  end

  def revoke(app_id)
    access_tokens = current_user.oauth_access_tokens.where(application_id: app_id)
    access_tokens.each(&:revoke)
    redirect_to :back, notice: t("messages.apps.disconnected")
  end
end
