class TwitterService
  def initialize(user)
    @user = user
  end

  def provider
    @provider ||= @user.providers.find_by(name: 'twitter')
  end

  def client
    return nil if provider.blank?

    @client ||= Twitter::REST::Client.new do |config|
      config.consumer_key        = ENV['TWITTER_CONSUMER_KEY']
      config.consumer_secret     = ENV['TWITTER_CONSUMER_SECRET']
      config.access_token        = provider.token
      config.access_token_secret = provider.token_secret
    end
  end

  def uids
    return [] if client.blank?

    client.friend_ids.to_a.map(&:to_s)
  end

  def share!(checkin)
    if checkin.twitter_url_hash.blank?
      checkin.update_column(:twitter_url_hash, checkin.generate_url_hash)
    end

    client.update(tweet_body(checkin))
  end

  private

  def tweet_body(checkin)
    data = {
      work_title: checkin.work.title,
      episode_number: get_episode_number(checkin),
      episode_title: get_episode_title(checkin),
      share_url: "#{ENV['ANNICT_HOST_SHORT']}/r/tw/#{checkin.twitter_url_hash}",
      share_hashtag: get_share_hashtag(checkin),
      comment: checkin.comment
    }

    body = generate_tweet_body(data)
    data = adjust_tweet_body(body, data)
    generate_tweet_body(data)
  end

  def get_episode_number(checkin)
    checkin.episode.single? ? "" : checkin.episode.number
  end

  def get_episode_title(checkin)
    if checkin.episode.single?
      ""
    else
      checkin.episode.title.presence || ""
    end
  end

  def get_share_hashtag(checkin)
    return "" if checkin.work.twitter_hashtag.blank?

    " ##{checkin.work.twitter_hashtag}"
  end

  def generate_tweet_body(data)
    if data[:comment].present?
      "#{data[:comment]}／#{data[:work_title]} #{data[:episode_number]}" +
        "を見ました #{data[:share_url]}#{data[:share_hashtag]}"
    else
      episode_title = data[:episode_title].present? ? "「#{data[:episode_title]}」" : ""

      "#{data[:work_title]} #{data[:episode_number]}#{episode_title}" +
        "を見ました #{data[:share_url]}#{data[:share_hashtag]}"
    end
  end

  def adjust_tweet_body(body, data)
    # 話数表記にシャープ記号が入っているとき (例: #1)、
    # 後続のテキストとくっついてハッシュタグにならないよう半角スペースを追加する
    data[:episode_number] += " " if data[:episode_number].include?("#")

    if body.length > 140
      if data[:comment].present?
        data[:comment] = truncate_text(data[:comment], body)
      else
        data[:episode_title] = truncate_text(data[:episode_title], body)
        body = generate_tweet_body(data)

        if body.length > 140
          data[:work_title] = truncate_text(data[:work_title], body)
        end
      end
    end

    data
  end

  def truncate_text(text, body)
    text_length = text.length - (body.length - 140)
    text.truncate(text_length)
  end
end
