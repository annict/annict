# frozen_string_literal: true

module Api::Internal
  class ReceivedChannelsController < Api::Internal::ApplicationController
    before_action :authenticate_user!

    def index
      render(json: current_user.receptions.pluck(:channel_id))
    end
  end
end
