class TwitterService
  def initialize(user)
    @user = user
  end

  def provider
    @provider ||= @user.providers.find_by(name: 'twitter')
  end

  def client
    @client ||= Twitter::REST::Client.new do |config|
      config.consumer_key        = ENV['TWITTER_CONSUMER_KEY']
      config.consumer_secret     = ENV['TWITTER_CONSUMER_SECRET']
      config.access_token        = provider.token
      config.access_token_secret = provider.token_secret
    end
  end

  def uids
    client.friend_ids.to_a.map(&:to_s)
  end
end
