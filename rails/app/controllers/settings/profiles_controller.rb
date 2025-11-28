# typed: false
# frozen_string_literal: true

module Settings
  class ProfilesController < ApplicationV6Controller
    before_action :authenticate_user!

    def update
      if current_user.profile.update(profile_params)
        flash[:notice] = t "messages._common.updated"
        redirect_to settings_profile_path
      else
        render :show, status: :unprocessable_entity
      end
    end

    private

    def profile_params
      params.require(:profile).permit(:image, :background_image, :description, :name, :url)
    end
  end
end
