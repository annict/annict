# frozen_string_literal: true

namespace :debug do
  task send_emails: :environment do
    user = User.find_by(username: "shimbaco")
    episode_record = user.episode_records.published.order(id: :desc).first

    EmailNotificationMailer.followed_user(user.id, user.id).deliver_now
    EmailNotificationMailer.liked_episode_record(user.id, user.id, episode_record.id).deliver_now
    EmailNotificationMailer.friends_joined(user.id, :twitter, [user.id]).deliver_now
    EmailNotificationMailer.favorite_works_added(user.id, episode_record.work_id).deliver_now
    EmailNotificationMailer.related_works_added(user.id, episode_record.work_id).deliver_now
  end
end
