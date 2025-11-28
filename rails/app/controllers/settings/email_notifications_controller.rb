# typed: false
# frozen_string_literal: true

module Settings
  class EmailNotificationsController < ApplicationV6Controller
    before_action :authenticate_user!, only: %i[show update]

    def show
      @email_notification = current_user.email_notification
    end

    def update
      @email_notification = current_user.email_notification

      if @email_notification.update(email_notification_params)
        flash[:notice] = t("messages.settings.email_notifications.updated")
        redirect_to settings_email_notification_path
      else
        render :show
      end
    end

    def unsubscribe
      return redirect_to root_path if params[:key].blank? || params[:action_name].blank?

      column = "event_#{params[:action_name]}"
      notification = EmailNotification.find_by!(unsubscription_key: params[:key])

      unless notification.has_attribute?(column)
        raise ActionController::RoutingError, "Not Found"
      end

      notification.update!(column => false)
    end

    private

    def email_notification_params
      params.require(:email_notification).permit(
        :event_followed_user, :event_liked_episode_record,
        :event_favorite_works_added, :event_related_works_added
      )
    end
  end
end
