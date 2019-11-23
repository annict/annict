# frozen_string_literal: true

namespace :user do
  task send_friends_joined_notification_email: :environment do
    Rails.logger = Logger.new(STDOUT) if Rails.env.development?
    Rails.logger.info "user:send_friends_joined_notification_email >> task started"

    %i(twitter facebook).each do |provider_name|
      user_ids = User.yesterday.select { |u| u.send(provider_name).present? }.pluck(:id)
      if user_ids.blank?
        Rails.logger.info "user:send_friends_joined_notification_email >> no users signed up via #{provider_name}"
        next
      end

      User.find_each do |u|
        Rails.logger.info "user:send_friends_joined_notification_email >> user: #{u.id}"

        unless u.email_notification.event_friends_joined?
          Rails.logger.info "user:send_friends_joined_notification_email >> user: #{u.id} >> stopped sending email"
          next
        end

        begin
          friend_ids = u.social_friends.users_via(provider_name).pluck(:id)
          Rails.logger.info "user:send_friends_joined_notification_email >> user: #{u.id} >> friend_ids: #{friend_ids.join(', ')}"
        rescue => err
          Rails.logger.info "user:send_friends_joined_notification_email >> user: #{u.id} >> failed to fetch friends. #{err}}"
          next
        end
        new_friend_ids = (user_ids & friend_ids) - u.followings.without_deleted.pluck(:id)
        Rails.logger.info "user:send_friends_joined_notification_email >> user: #{u.id} >> new_friend_ids: #{new_friend_ids.join(', ')}"

        if new_friend_ids.blank?
          Rails.logger.info "user:send_friends_joined_notification_email >> user: #{u.id} >> no new friends found"
          next
        end

        Rails.logger.info "user:send_friends_joined_notification_email >> user: #{u.id} >> queuing email..."
        EmailNotificationService.send_email(
          "friends_joined",
          u,
          provider_name.to_s,
          new_friend_ids
        )
        Rails.logger.info "user:send_friends_joined_notification_email >> user: #{u.id} >> queued"
      end
    end

    Rails.logger.info "user:send_friends_joined_notification_email >> task processed"
  end
end
