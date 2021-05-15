# frozen_string_literal: true

class ProvidersController < ApplicationController
  before_action :authenticate_user!

  def destroy
    provider = current_user.providers.find(params[:id])

    ActiveRecord::Base.transaction do
      case provider.name
      when "twitter"
        provider.user.setting.update_column(:share_record_to_twitter, false)
      end

      provider.destroy
    end

    flash[:notice] = t("messages.providers.removed")
    redirect_back fallback_location: providers_path
  end
end
