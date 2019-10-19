# frozen_string_literal: true

class ProfilesController < ApplicationController
  before_action :authenticate_user!

  def update
    if current_user.profile.update(profile_params)
      redirect_to profile_path, notice: t("messages.profiles.saved")
    else
      render :show
    end
  end
  
  private

  def profile_params
    params.require(:profile).permit(:image, :background_image, :description, :name, :url)
  end
end
