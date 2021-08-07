# frozen_string_literal: true

namespace :debug do
  task send_emails: :environment do
    user = User.find_by(username: "shimbaco")
    episode_record = user.episode_records.only_kept.order(id: :desc).first
    works = Anime.only_kept.joins(:casts).where(casts: {character_id: user.favorite_characters.pluck(:id)}).limit(3)

    EmailNotificationMailer.followed_user(user.id, user.id).deliver_now
    EmailNotificationMailer.liked_episode_record(user.id, user.id, episode_record.id).deliver_now
    EmailNotificationMailer.favorite_works_added(user.id, works.pluck(:id)).deliver_now
    EmailNotificationMailer.related_works_added(user.id, works.pluck(:id)).deliver_now

    ForumMailer.comment_notification(user.id, ForumComment.first.id).deliver_now

    AdminMailer.episode_created_notification(Episode.first.id).deliver_now
    AdminMailer.error_in_episode_generator_notification(Slot.first.id, "テストなんだよなあ").deliver_now
  end
end
