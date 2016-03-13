# frozen_string_literal: true

class OptionsController < ApplicationController
  permits :hide_checkin_comment, model_name: "Setting"

  before_action :authenticate_user!

  def index
    render layout: "v1/application"
  end

  def update(setting)
    if current_user.setting.update_attributes(setting)
      redirect_to options_path, notice: "設定を更新しました"
    else
      render :show, layout: "v1/application"
    end
  end
end
