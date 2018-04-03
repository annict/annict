# frozen_string_literal: true

module Settings
  class EmailNotificationsController < ApplicationController
    permits :event_followed_user, :event_liked_record, :event_friends_joined,
      :event_favorite_works_added

    before_action :authenticate_user!, only: %i(show update)

    def show
      @email_notification = current_user.email_notification
    end

    def update(email_notification)
      @email_notification = current_user.email_notification

      if @email_notification.update(email_notification)
        flash[:notice] = t("messages.settings.email_notifications.updated")
        redirect_to settings_email_notification_path
      else
        render :show
      end
    end

    def unsubscribe(key: nil, action_name: nil)
      return redirect_to root_path if key.blank? || action_name.blank?

      column = "event_#{action_name}"
      notification = EmailNotification.find_by!(unsubscription_key: key)

      unless notification.has_attribute?(column)
        raise ActionController::RoutingError, "Not Found"
      end

      notification.update_column(column, false)
    end
  end
end
