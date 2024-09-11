# typed: false
# frozen_string_literal: true

class Deprecated::EmailNotificationService
  def self.send_email(action, user, *)
    return unless user.email_notification.send("event_#{action}?")
    new.send_email(action, user.id, *)
  end

  def send_email(action, user_id, *)
    EmailNotificationMailer.send(action.to_s, user_id, *).deliver_later(priority: 20)
  end
end
