# frozen_string_literal: true

class OptionsController < ApplicationController
  permits :hide_record_comment, :hide_supporter_badge, model_name: "Setting"

  before_action :authenticate_user!

  def update(setting)
    if current_user.setting.update_attributes(setting)
      redirect_to options_path, notice: t("messages._common.updated")
    else
      render :show
    end
  end
end
