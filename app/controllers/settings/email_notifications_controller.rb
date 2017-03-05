# frozen_string_literal: true

module Settings
  class EmailNotificationsController < ApplicationController
    def unsubscribe(key, action_name)
      column = "event_#{action_name}"
      notification = EmailNotification.find_by!(unsubscription_key: key)

      unless notification.has_attribute?(column)
        raise ActionController::RoutingError.new("Not Found")
      end

      notification.update_column(column, false)
    end
  end
end
