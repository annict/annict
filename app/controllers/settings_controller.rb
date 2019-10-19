# frozen_string_literal: true

class SettingsController < ApplicationController
  before_action :authenticate_user!

  def index
    return render(:index) unless device_pc?
    redirect_to profile_path
  end
end
