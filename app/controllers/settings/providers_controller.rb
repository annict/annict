# frozen_string_literal: true

module Settings
  class ProvidersController < ApplicationV6Controller
    before_action :authenticate_user!

    def destroy
      provider = current_user.providers.find(params[:provider_id])

      ActiveRecord::Base.transaction do
        case provider.name
        when "twitter"
          provider.user.setting.update!(share_record_to_twitter: false)
        end

        provider.destroy
      end

      flash[:notice] = t("messages.providers.removed")
      redirect_back fallback_location: settings_provider_list_path
    end
  end
end
