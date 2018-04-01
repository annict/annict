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

    @client ||= Twitter::REST::Client.new do |config|
      config.consumer_key = ENV.fetch("TWITTER_CONSUMER_KEY")
      config.consumer_secret = ENV.fetch("TWITTER_CONSUMER_SECRET")
      config.access_token = provider.token
      config.access_token_secret = provider.token_secret
    end
  end

  def uids
    client.friend_ids.to_a.map(&:to_s)
  rescue Twitter::Error::Unauthorized
    @user.expire_twitter_token
    []
  end

  def share!(record, data)
    record.update_column(:twitter_url_hash, record.generate_url_hash) if record.twitter_url_hash.blank?

    data[:share_url] = "#{record.user.annict_host}/r/tw/#{record.twitter_url_hash}"

    begin
      client.update(tweet_body(data))
    rescue Twitter::Error::Unauthorized
      record.user.expire_twitter_token
    end
  end

  def share_review!(review)
    client.update(review.decorate.tweet_body)
  rescue Twitter::Error::Unauthorized
    review.user.expire_twitter_token
  end

  def share_status!(status)
    client.update(status.decorate.tweet_body)
  rescue Twitter::Error::Unauthorized
    status.user.expire_twitter_token
  end

  private

  def tweet_body(data)
    body = generate_tweet_body(data)
    data = adjust_tweet_body(body, data)
    generate_tweet_body(data)
  end

  def generate_tweet_body(data)
    if data[:locale] == "ja"
      base_body = "#{data[:work_title]} #{data[:episode_number]}を見ました #{data[:share_url]}#{data[:share_hashtag]}"
      return base_body if data[:comment].blank?
      "#{data[:comment]}／#{base_body}"
    else
      base_body = "Watched: #{data[:work_title]} #{data[:episode_number]} #{data[:share_url]}#{data[:share_hashtag]}"
      return base_body if data[:comment].blank?
      "#{data[:comment]} / #{base_body}"
    end
  end

  def adjust_tweet_body(body, data)
    # 話数表記にシャープ記号が入っているとき (例: #1)、
    # 後続のテキストとくっついてハッシュタグにならないよう半角スペースを追加する
    data[:episode_number] += " " if data[:episode_number]&.include?("#")

    if body.length > 140
      data[:comment] = truncate_text(data[:comment], body) if data[:comment].present?
      body = generate_tweet_body(data)
      data[:work_title] = truncate_text(data[:work_title], body) if body.length > 140
    end

    data
  end

  def truncate_text(text, body)
    text_length = text.length - (body.length - 140)
    text.truncate(text_length)
  end
end
