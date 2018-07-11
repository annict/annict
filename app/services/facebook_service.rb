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
end
