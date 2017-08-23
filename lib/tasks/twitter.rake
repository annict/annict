# frozen_string_literal: true

namespace :twitter do
  task post_tweets_to_slack: :environment do
    Rails.logger = Logger.new(STDOUT) if Rails.env.development?
    Rails.logger.info "twitter:post_tweets_to_slack >> Task started"

    provider = Provider.joins(:user).where(users: { username: "shimbaco_anime" }, name: "twitter").first
    client ||= Twitter::REST::Client.new do |config|
      config.consumer_key = ENV.fetch("TWITTER_CONSUMER_KEY")
      config.consumer_secret = ENV.fetch("TWITTER_CONSUMER_SECRET")
      config.access_token = provider.token
      config.access_token_secret = provider.token_secret
    end

    Work.where.not(twitter_username: [nil, ""]).find_each do |w|
      Rails.logger.info "twitter:post_tweets_to_slack >> Creating TwitterUser: #{w.twitter_username}"
      TwitterUser.where(screen_name: w.twitter_username).first_or_create(work: w)
    end

    screen_names = TwitterUser.published.where(followed_at: nil).pluck(:screen_name)
    Rails.logger.info "twitter:post_tweets_to_slack >> Following screen_names: #{screen_names.join(', ')}"
    users = client.follow(screen_names)
    users.each do |u|
      twitter_user = TwitterUser.where("lower(screen_name) = ?", u.screen_name.downcase).first

      if twitter_user.present?
        Rails.logger.info "twitter:post_tweets_to_slack >> Updating twitter_user: #{twitter_user.screen_name}"
        twitter_user.update(user_id: u.id, followed_at: Time.now)
      else
        Rails.logger.info "twitter:post_tweets_to_slack >> Can't find TwitterUser: #{u.screen_name.downcase}"
      end
    end

    options = {
      exclude_replies: true,
      include_rts: false
    }
    latest_tweet = TwitterTweet.order(tweet_id: :desc).first
    options[:since_id] = latest_tweet.tweet_id if latest_tweet.present?
    Rails.logger.info "twitter:post_tweets_to_slack >> Fetching timeline with this options: #{options}"
    tweets = client.home_timeline(options).sort_by(&:id)

    tweets.each do |t|
      Rails.logger.info "twitter:post_tweets_to_slack >> Tweet: #{t.id}"

      if t.user.screen_name == "shimbaco_anime"
        Rails.logger.info "twitter:post_tweets_to_slack >> Tweet: #{t.id} >> This tweet is mine. Skipped"
        next
      end

      user = TwitterUser.find_by(user_id: t.user.id)
      if user.blank?
        user = TwitterUser.find_by(screen_name: t.user.screen_name)
        if user.present?
          user.update_column(:user_id, t.user.id)
        else
          Rails.logger.info "twitter:post_tweets_to_slack >> Tweet: #{t.id} >> User not found: #{t.user.id}"
        end
      end

      attrs = {
        text: t.text,
        user_screen_name: t.user.screen_name,
        user_name: t.user.name
      }
      tweet = TwitterTweet.where(tweet_id: t.id, twitter_user: user).first_or_create!(attrs)

      Rails.logger.info "twitter:post_tweets_to_slack >> Posting to Slack: #{tweet.id}"
      tweet.notify_slack
    end
  end
end
