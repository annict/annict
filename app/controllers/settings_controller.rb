class SettingsController < ApplicationController
  before_action :authenticate_user!

  def index
    return render(:index) if browser.mobile?
    redirect_to profile_path
  end
end
