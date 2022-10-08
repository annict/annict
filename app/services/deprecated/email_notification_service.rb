# frozen_string_literal: true

class Deprecated::EmailNotificationService
  def self.send_email(action, user, *args)
    return unless user.email_notification.send("event_#{action}?")
    new.send_email(action, user.id, *args)
  end

  def send_email(action, user_id, *args)
    EmailNotificationMailer.send(action.to_s, user_id, *args).deliver_later(priority: 20)
  end
end
