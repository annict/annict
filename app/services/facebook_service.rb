# frozen_string_literal: true

class FacebookService
  def initialize(user)
    @user = user
  end

  def provider
    @provider ||= @user.providers.find_by(name: "facebook")
  end

  def client
    @client ||= Koala::Facebook::API.new(provider.token)
  end

  def uids
    client.get_connections(:me, :friends).map { |friend| friend["id"] }
  end

  def share!(record)
    if record.facebook_url_hash.blank?
      record.update_column(:facebook_url_hash, record.generate_url_hash)
    end

    title = record.work.title
    episode_title = record.episode.decorate.title_with_number
    title += " #{episode_title}" if title != episode_title
    message = record.comment.squish if record.comment.present?
    link = if Rails.env.development?
      "https://annict.com/r/fb/#{record.facebook_url_hash}"
    else
      "#{record.user.annict_url}/r/fb/#{record.facebook_url_hash}"
    end
    caption = "Annict | アニクト - 見たアニメを記録して、共有しよう"

    work_image = record.work.work_image
    source = if work_image.present? && Rails.env.production?
      # Using `rescue` to avoid occurring error `undefined method `ann_image_url'`
      work_image.decorate.image_url(:attachment, size: "600x315") rescue nil # rubocop:disable Style/RescueModifier
    end
    source = "https://annict.com/images/og_image.png" if source.blank?

    client.put_connections("me", "feed",
      name: title,
      message: message,
      link: link,
      caption: caption,
      source: source)
  end

  def share_review!(review, image_url)
    client.put_connections("me", "feed",
      name: review.work.title,
      message: review.body,
      link: review.decorate.detail_url,
      caption: "Annict | アニクト - 見たアニメを記録して、共有しよう",
      source: image_url)
  end

  def share_status!(status, image_url)
    utm = {
      utm_source: "facebook",
      utm_medium: "status_share",
      utm_campaign: status.user.username
    }

    client.put_connections("me", "feed",
      name: status.work.title,
      message: status.decorate.facebook_share_body,
      link: "#{status.decorate.library_url}?#{utm.to_query}",
      caption: "Annict | アニクト - 見たアニメを記録して、共有しよう",
      source: image_url)
  end
end
