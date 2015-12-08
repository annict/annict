class ProfilesController < ApplicationController
  permits :tombo_avatar, :tombo_background_image, :description, :name

  before_action :authenticate_user!

  def update(profile)
    if current_user.profile.update_attributes(profile)
      redirect_to profile_path, notice: t("profiles.saved")
    else
      render :show
    end
  end
end
