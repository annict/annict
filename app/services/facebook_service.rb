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

  def share!(resource, image_url)
    client.put_connections("me", "feed",
      name: resource.facebook_share_title,
      message: resource.facebook_share_body,
      link: resource.share_url_with_query(:facebook),
      caption: "Annict | アニクト - 見たアニメを記録して、共有しよう",
      source: image_url)
  end
end
