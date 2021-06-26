# frozen_string_literal: true

class SettingsController < ApplicationV6Controller
  before_action :authenticate_user!

  def index
    return render(:index) unless device_pc?
    redirect_to settings_profile_path
  end
end
