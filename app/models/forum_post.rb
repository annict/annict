# frozen_string_literal: true

# == Schema Information
#
# Table name: forum_posts
#
#  id                   :integer          not null, primary key
#  user_id              :integer          not null
#  forum_category_id    :integer          not null
#  title                :string           not null
#  body                 :text             default(""), not null
#  forum_comments_count :integer          default(0), not null
#  edited_at            :datetime
#  created_at           :datetime         not null
#  updated_at           :datetime         not null
#  last_commented_at    :datetime         not null
#
# Indexes
#
#  index_forum_posts_on_forum_category_id  (forum_category_id)
#  index_forum_posts_on_user_id            (user_id)
#

class ForumPost < ApplicationRecord
  belongs_to :user
  belongs_to :forum_category
  has_many :forum_comments, dependent: :destroy
  has_many :forum_post_participants, dependent: :destroy

  validates :body, presence: true, length: { maximum: 5000 }
  validates :forum_category, presence: true
  validates :last_commented_at, presence: true
  validates :title, presence: true, length: { maximum: 100 }
  validates :user, presence: true

  def notify_discord
    return unless Rails.env.production?

    options = {
      url: ENV.fetch("ANNICT_DISCORD_WEBHOOK_URL_FOR_FORUM-#{forum_category.slug.tr('_', '-').upcase}")
    }
    host = ENV.fetch("ANNICT_URL")
    url = Rails.application.routes.url_helpers.forum_post_url(self, host: host)
    message = "@everyone #{user.profile.name} created the post #{title} #{url}"

    Discord::Notifier.message(message, options)
  end
end
