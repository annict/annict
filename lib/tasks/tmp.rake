# frozen_string_literal: true

namespace :tmp do
  task create_email_notifications: :environment do
    User.find_each do |u|
      next if u.email_notification.present?
      puts "user: #{u.id}"
      unsubscription_key = "#{SecureRandom.uuid}-#{SecureRandom.uuid}"
      u.create_email_notification!(unsubscription_key: unsubscription_key)
    end
  end
end
