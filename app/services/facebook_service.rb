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

  def share!(checkin, controller)
    if checkin.facebook_url_hash.blank?
      checkin.update_column(:facebook_url_hash, checkin.generate_url_hash)
    end

    title = checkin.work.title
    episode_title = checkin.episode.decorate.title_with_number
    title += " #{episode_title}" if title != episode_title
    message = checkin.comment.squish if checkin.comment.present?
    link = if Rails.env.development?
             "http://www.annict.com/r/fb/#{checkin.facebook_url_hash}"
           else
             "#{ENV['ANNICT_URL']}/r/fb/#{checkin.facebook_url_hash}"
           end
    caption = "Annict | アニクト - 見たアニメを記録して、共有しよう"
    view_context = controller.view_context
    source = view_context.annict_image_url(checkin.work.item, :tombo_image, size: "600x315")

    client.delay.put_connections("me", "feed",
      name: title,
      message: message,
      link: link,
      caption: caption,
      source: source
    )
  end
end
