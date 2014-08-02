class SettingsController < ApplicationController
  permits :avatar, :background_image, :description, :name, model_name: 'Profile'

  before_filter :authenticate_user!

  def edit
    @profile = current_user.profile
  end

  def update(profile)
    @profile = current_user.profile

    if @profile.update_attributes(profile)
      redirect_to :back, notice: t('profiles.saved')
    else
      render 'edit'
    end
  end
end