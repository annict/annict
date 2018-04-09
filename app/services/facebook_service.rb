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

  def share_record!(record, image_url)
    utm = {
      utm_source: "facebook",
      utm_medium: "record_share",
      utm_campaign: record.user.username
    }
    title = record.work.title
    title += " #{record.episode.decorate.title_with_number}" if record.episode.present?

    client.put_connections("me", "feed",
      name: title,
      message: record.decorate.facebook_share_body,
      link: "#{record.detail_url}?#{utm.to_query}",
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
      link: "#{status.library_url}?#{utm.to_query}",
      caption: "Annict | アニクト - 見たアニメを記録して、共有しよう",
      source: image_url)
  end
end
