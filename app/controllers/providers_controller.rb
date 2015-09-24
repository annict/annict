class ProvidersController < ApplicationController
  before_action :authenticate_user!

  def destroy(id)
    provider = current_user.providers.find(id)

    ActiveRecord::Base.transaction do
      case provider.name
      when "twitter"
        provider.user.setting.update_column(:share_record_to_twitter, false)
      when "facebook"
        provider.user.setting.update_column(:share_record_to_facebook, false)
      end

      provider.destroy
    end

    redirect_to :back, notice: "連携を解除しました"
  end
end
