# typed: false
# frozen_string_literal: true

module Settings
  class ProvidersController < ApplicationV6Controller
    before_action :authenticate_user!

    def destroy
      provider = current_user.providers.find(params[:provider_id])
      provider.destroy

      flash[:notice] = t("messages.providers.removed")
      redirect_back fallback_location: settings_provider_list_path
    end
  end
end
