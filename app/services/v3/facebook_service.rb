# frozen_string_literal: true

module V3
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
      []
    end
  end
end
