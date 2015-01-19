class TwitterCheckinShareWorker
  include Sidekiq::Worker

  def perform(checkin_id)
    @rest = 0
    @checkin = Checkin.find(checkin_id)

    if @checkin.twitter_url_hash.blank?
      @checkin.update_column(:twitter_url_hash, @checkin.generate_url_hash)
    end

    twitter = @checkin.user.providers.where(name: 'twitter').first
    client = Twitter::REST::Client.new do |config|
      config.consumer_key        = ENV['TWITTER_CONSUMER_KEY']
      config.consumer_secret     = ENV['TWITTER_CONSUMER_SECRET']
      config.access_token        = twitter.token
      config.access_token_secret = twitter.token_secret
    end

    client.update(tweet_text(@checkin.comment))
  end


  private

  def get_work_title
    max_length = 35
    title = @checkin.work.title

    if @checkin.comment.present?
      if max_length < title.length
        title.truncate(max_length)
      else
        @rest += (max_length - title.length)
        title
      end
    else
      title.truncate(max_length)
    end
  end

  def get_episode_number
    max_length = 10
    number = @checkin.episode.single? ? '' : @checkin.episode.number

    if max_length < number.length
      number.truncate(max_length)
    else
      @rest += (max_length - number.length)
      number
    end
  end

  def get_episode_title
    title = if @checkin.episode.single?
      ''
    else
      @checkin.episode.title.presence || ''
    end

    if @checkin.comment.present?
      @rest += 10
    else
      max_length = 30
      title = if max_length < title.length
        title.truncate(max_length)
      else
        title
      end

      title.present? ? "「#{title}」" : ' '
    end
  end

  def get_share_url
    "#{ENV['ANNICT_URL']}/r/tw/#{@checkin.twitter_url_hash}"
  end

  def get_share_hashtag
    max_length = 20
    hashtag = @checkin.work.twitter_hashtag.presence || ''

    if max_length < hashtag.length
      hashtag.truncate(max_length)
    else
      @rest += (max_length - hashtag.length)
      hashtag
    end
  end

  def tweet_text(comment)
    max_length = 10

    work_title     = get_work_title
    episode_number = get_episode_number
    episode_title  = get_episode_title
    share_url      = get_share_url
    share_hashtag  = get_share_hashtag

    if comment.present?
      comment_length = max_length + @rest
      "#{comment.truncate(comment_length)} / #{work_title} #{episode_number} にチェックイン！#{share_url} #{share_hashtag}"
    else
      "#{work_title} #{episode_number}#{episode_title}にチェックイン！#{share_url} #{share_hashtag}"
    end
  end
end
