class FacebookService
  def initialize(user)
    @user = user
  end

  def provider
    @provider ||= @user.providers.find_by(name: 'facebook')
  end

  def client
    @client ||= Koala::Facebook::API.new(provider.token)
  end

  def uids
    client.get_connections(:me, :friends).map { |friend| friend['id'] }
  end

  def share!(checkin)
    if checkin.facebook_url_hash.blank?
      checkin.update_column(:facebook_url_hash, checkin.generate_url_hash)
    end

    work_title = checkin.episode.work.title
    episode_number_title = checkin.episode.number_title
    title = "#{work_title} #{episode_number_title}".truncate(30)

    message = checkin.comment.squish.truncate(50) if checkin.comment.present?
    link = if Rails.env.development?
             "http://www.annict.com/checkins/redirect/fb/#{checkin.facebook_url_hash}"
           else
             "#{ENV['ANNICT_URL']}/checkins/redirect/fb/#{checkin.facebook_url_hash}"
           end
    name = I18n.t("checkins.facebook_share_text", title: title)
    caption = "Checkined by #{checkin.user.username}"

    client.put_connections("me", "feed",
      message:      message,
      link:         link,
      name:         name,
      caption:      caption
    )
  end
end
