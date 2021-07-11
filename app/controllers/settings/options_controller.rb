# frozen_string_literal: true

module Settings
  class OptionsController < ApplicationV6Controller
    before_action :authenticate_user!

    def update
      if current_user.setting.update(setting_params)
        redirect_to settings_option_list_path, notice: t("messages._common.updated")
      else
        render :show, status: :unprocessable_entity
      end
    end

    private

    def setting_params
      params.require(:setting).permit(:hide_record_body, :hide_supporter_badge, :share_status_to_twitter)
    end
  end
end
