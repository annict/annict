# frozen_string_literal: true

module V3
  class SettingsController < V3::ApplicationController
    before_action :authenticate_user!

    def index
      return render(:index) unless device_pc?
      redirect_to settings_profile_path
    end
  end
end
