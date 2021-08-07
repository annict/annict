# frozen_string_literal: true

namespace :twitter do
  task fetch_tweets_from_official_accounts: :environment do
    client = Twitter::REST::Client.new { |config|
      config.consumer_key = ENV.fetch("TWITTER_CONSUMER_KEY_FOR_DEV")
      config.consumer_secret = ENV.fetch("TWITTER_CONSUMER_SECRET_FOR_DEV")
      config.access_token = ENV.fetch("TWITTER_ACCESS_TOKEN_FOR_DEV")
      config.access_token_secret = ENV.fetch("TWITTER_ACCESS_SECRET_FOR_DEV")
    }
    lists = TwitterWatchingList.all
    channels = Channel.only_kept.joins(:channel_group).merge(ChannelGroup.only_kept)
    channel_names = channels.pluck(:name) + channels.pluck(:name_alter).map { |name| name.split(",") }.flatten

    lists.each do |list|
      puts "--- list: @#{list.username}/#{list.name}"

      options = {
        count: 100,
        tweet_mode: "extended" # Get tweet which is not truncated https://github.com/sferik/twitter/issues/813
      }
      options[:since_id] = list.since_id if list.since_id
      tweets = client.list_timeline(list.username, list.name, options)

      tweets.reverse_each do |tweet|
        next if tweet.retweet? || tweet.reply?

        puts "tweet: #{tweet.url}"

        # Prevent to embed https://support.discordapp.com/hc/en-us/articles/206342858--How-do-I-disable-auto-embed-
        tweet_body = tweet.attrs[:full_text].gsub(%r{(https?:\/\/[\S]+)}, "<\\1>")
        work_urls = Work
          .only_kept
          .where(twitter_username: tweet.user.screen_name)
          .order_by_season(:desc)
          .limit(3)
          .select(:title, :id)
          .map { |w|
            "- #{w.title}: <https://annict.jp/db/works/#{w.id}/edit>"
          }.join("\n")

        # Skip posting to Discord if tweet body does not contain channel name
        next unless channel_names.any? { |name| tweet_body.include?(name) }

        Discord::Notifier.message(
          "#{tweet_body}\nTweet URL:\n- <#{tweet.url}>\nAnnict DB:\n#{work_urls}\n---",
          username: tweet.user.name,
          avatar_url: tweet.user.profile_image_uri_https&.to_s,
          url: list.discord_webhook_url,
          wait: true
        )
      end

      latest_tweet = tweets.first
      if latest_tweet
        puts "latest tweet: #{latest_tweet.url}"
        list.update_column(:since_id, latest_tweet.id.to_s)
      end
    end
  end
end
