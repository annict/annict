class ApplicationController < ActionController::Base
  # Prevent CSRF attacks by raising an exception.
  # For APIs, you may want to use :null_session instead.
  protect_from_forgery with: :exception

  before_filter :store_flash_message


  private

  def after_sign_in_path_for(resource)
    if 'Staff' == resource.class.name
      marie_root_path
    else
      root_path
    end
  end

  def after_sign_out_path_for(resource_or_scope)
    if :staff == resource_or_scope
      new_staff_session_path
    else
      root_path
    end
  end

  def set_work
    @work = Work.find(params[:work_id])
  end

  def set_episode
    @episode = @work.episodes.find(params[:episode_id])
  end

  def set_checkin
    @checkin = @episode.checkins.find(params[:checkin_id])
  end

  def store_flash_message
    key = flash.keys.first
    message = { type: key.to_s, body: flash[key] } if flash[key].present?

    gon.push(flash: message.presence || {})
  end
end
