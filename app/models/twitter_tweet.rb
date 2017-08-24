# frozen_string_literal: true
# == Schema Information
#
# Table name: twitter_tweets
#
#  id               :integer          not null, primary key
#  twitter_user_id  :integer          not null
#  user_screen_name :string           not null
#  user_name        :string           not null
#  tweet_id         :string           not null
#  text             :text             not null
#  created_at       :datetime         not null
#  updated_at       :datetime         not null
#
# Indexes
#
#  index_twitter_tweets_on_tweet_id         (tweet_id) UNIQUE
#  index_twitter_tweets_on_twitter_user_id  (twitter_user_id)
#

class TwitterTweet < ApplicationRecord
  belongs_to :twitter_user

  def notify_slack
    return unless Rails.env.production?

    webhook_url = ENV.fetch("ANNICT_SLACK_WEBHOOK_URL_FOR_NOTIFICATIONS")
    options = {
      channel: "#official-twitter",
      username: "Notifier",
      icon_emoji: ":annict:"
    }

    messages = ["\n\n---\n"]

    if twitter_user.work.present?
      messages << "<https://annict.jp/db/works/#{twitter_user.work.id}/edit|Annict DB>"
    end

    messages << url

    notifier = Slack::Notifier.new(webhook_url, options)
    notifier.ping(messages.join("\n"))
  end

  def url
    "https://twitter.com/#{user_screen_name}/status/#{tweet_id}"
  end
end
