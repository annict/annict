# frozen_string_literal: true

class OptionsController < ApplicationController
  before_action :authenticate_user!

  def update
    if current_user.setting.update(setting_params)
      redirect_to options_path, notice: t("messages._common.updated")
    else
      render :show
    end
  end

  private

  def setting_params
    params.require(:setting).permit(:hide_record_body, :hide_supporter_badge, :share_status_to_twitter)
  end
end
