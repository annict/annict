# frozen_string_literal: true

module Api::Internal
  class ReceivedChannelsController < Api::Internal::ApplicationController
    def index
      return render(json: []) unless user_signed_in?

      render(json: current_user.receptions.pluck(:channel_id))
    end
  end
end
