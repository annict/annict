class SettingsController < ApplicationController
  permits :hide_checkin_comment

  before_action :authenticate_user!

  def update(setting)
    if current_user.setting.update_attributes(setting)
      redirect_to setting_path, notice: '設定を更新しました'
    else
      render '/settings/show'
    end
  end
end
