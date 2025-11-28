# typed: false
# frozen_string_literal: true

class WorkDisplayOptionsController < ApplicationV6Controller
  def show
    redirect_path = params[:to].presence || root_path

    display_options = Setting.display_option_work_list.values
    return redirect_back(fallback_location: send(redirect_path)) unless params[:display].in?(display_options)

    if user_signed_in? && current_user.setting.display_option_work_list != params[:display]
      current_user.setting.update_column(:display_option_work_list, params[:display])
    end

    redirect_to "#{redirect_path}?display=#{params[:display]}"
  end
end
