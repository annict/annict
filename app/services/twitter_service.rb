# frozen_string_literal: true

class TwitterService
  def initialize(user)
    @user = user
  end

  def provider
    @provider ||= @user.providers.find_by(name: "twitter")
  end

  def client
    return nil if provider.blank?

    @client ||= Twitter::REST::Client.new { |config|
      config.consumer_key = ENV.fetch("TWITTER_CONSUMER_KEY")
      config.consumer_secret = ENV.fetch("TWITTER_CONSUMER_SECRET")
      config.access_token = provider.token
      config.access_token_secret = provider.token_secret
    }
  end

  def uids
    client.friend_ids.to_a.map(&:to_s)
  rescue Twitter::Error::Unauthorized
    @user.expire_twitter_token
    []
  end

  def share!(resource)
    client&.update(resource.twitter_share_body)
  rescue Twitter::Error::Unauthorized
    resource.user.expire_twitter_token
  end
end
