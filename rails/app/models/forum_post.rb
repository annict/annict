# typed: false
# frozen_string_literal: true

class ForumPost < ApplicationRecord
  include UgcLocalizable

  belongs_to :user
  belongs_to :forum_category
  has_many :forum_comments, dependent: :destroy
  has_many :forum_post_participants, dependent: :destroy

  validates :body, presence: true, length: {maximum: 10_000}
  validates :forum_category, presence: true
  validates :last_commented_at, presence: true
  validates :title, presence: true, length: {maximum: 100}
  validates :user, presence: true

  def notify_discord
    return unless Rails.env.production?

    options = {
      url: ENV.fetch("ANNICT_DISCORD_WEBHOOK_URL_FOR_FORUM_#{forum_category.slug.upcase}")
    }
    host = ENV.fetch("ANNICT_URL")
    url = Rails.application.routes.url_helpers.forum_post_url(self, host: host)
    message = [
      user.profile.name,
      "created the post",
      title,
      url
    ].select(&:present?).join(" ")

    Discord::Notifier.message(message, options)
  end
end
