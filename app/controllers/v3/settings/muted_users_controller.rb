# frozen_string_literal: true

module V3::Settings
  class MutedUsersController < V3::ApplicationController
    before_action :authenticate_user!

    def index
      @mute_users = current_user.mute_users.order(id: :desc)
    end
  end
end
