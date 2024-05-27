# typed: false
# frozen_string_literal: true

module Chat
  class HomeController < Chat::ApplicationController
    def index
      redirect_to discord_invite_url
    end
  end
end
