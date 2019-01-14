# frozen_string_literal: true

namespace :twitter do
  task fetch_tweets_from_official_accounts: :environment do
    client = Twitter::REST::Client.new do |config|
      config.consumer_key = ENV.fetch("TWITTER_CONSUMER_KEY_FOR_DEV")
      config.consumer_secret = ENV.fetch("TWITTER_CONSUMER_SECRET_FOR_DEV")
      config.access_token = ENV.fetch("TWITTER_ACCESS_TOKEN_FOR_DEV")
      config.access_token_secret = ENV.fetch("TWITTER_ACCESS_SECRET_FOR_DEV")
    end
    lists = TwitterWatchingList.all

    lists.each do |list|
      puts "--- list: @#{list.username}/#{list.name}"

      options = {
        count: 100
      }
      options[:since_id] = list.since_id if list.since_id
      tweets = client.list_timeline(list.username, list.name, options)

      tweets.reverse.each do |tweet|
        next if tweet.retweet? || tweet.reply?
        puts "tweet: #{tweet.url}"

        # Prevent to embed https://support.discordapp.com/hc/en-us/articles/206342858--How-do-I-disable-auto-embed-
        tweet_body = tweet.full_text.gsub(%r{(https?:\/\/[\S]+)}, "<\\1>")
        Discord::Notifier.message(
          "#{tweet_body}\n<#{tweet.url}>",
          username: tweet.user.name,
          avatar_url: tweet.user.profile_image_uri_https&.to_s,
          url: list.discord_webhook_url
        )

        sleep 1
      end

      latest_tweet = tweets.first
      if latest_tweet
        puts "latest tweet: #{latest_tweet.url}"
        list.update_column(:since_id, latest_tweet.id.to_s)
      end
    end
  end
end
