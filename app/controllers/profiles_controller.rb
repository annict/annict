class ProfilesController < ApplicationController
  permits :avatar, :background_image, :description, :name

  before_filter :authenticate_user!

  def update(profile)
    if current_user.profile.update_attributes(profile)
      redirect_to setting_path, notice: t('profiles.saved')
    else
      render '/settings/show'
    end
  end
end
