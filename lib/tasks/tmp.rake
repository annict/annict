# frozen_string_literal: true

namespace :tmp do
  task copy_twitter_username_to_twitter_users: :environment do
    Work.where.not(twitter_username: [nil, ""]).find_each do |w|
      puts "w.twitter_username: #{w.twitter_username}"
      TwitterUser.where(screen_name: w.twitter_username).first_or_create(work: w)
    end
  end

  task follow: :environment do
    provider = Provider.joins(:user).where(users: { username: "shimbaco_anime" }, name: "twitter").first
    client ||= Twitter::REST::Client.new do |config|
      config.consumer_key = ENV.fetch("TWITTER_CONSUMER_KEY")
      config.consumer_secret = ENV.fetch("TWITTER_CONSUMER_SECRET")
      config.access_token = provider.token
      config.access_token_secret = provider.token_secret
    end

    (1..1000).each do |i|
      puts "page: #{i}"
      screen_names = TwitterUser.published.where(followed_at: nil).page(i).per(100).pluck(:screen_name)
      puts "twitter:post_tweets_to_slack >> Following screen_names: #{screen_names.join(', ')}"
      users = client.follow(screen_names)
      users.each do |u|
        puts "twitter:post_tweets_to_slack >> u.screen_name.downcase: #{u.screen_name.downcase}"
        twitter_user = TwitterUser.where("lower(screen_name) = ?", u.screen_name.downcase).first
        next if twitter_user.blank?
        puts "twitter:post_tweets_to_slack >> Updating twitter_user: #{twitter_user.screen_name}"
        twitter_user.update(user_id: u.id, followed_at: Time.now)
      end

      sleep 70
    end
  end
end
