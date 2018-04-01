# frozen_string_literal: true

class OptionsController < ApplicationController
  permits :hide_record_comment, :hide_supporter_badge, :share_status_to_facebook, :share_status_to_twitter,
          model_name: "Setting"

  before_action :authenticate_user!

  def update(setting)
    if current_user.setting.update(setting)
      redirect_to options_path, notice: t("messages._common.updated")
    else
      render :show
    end
  end
end
