# frozen_string_literal: true

module My
  class ReceiveChannelButtonsController < My::ApplicationController
    layout false

    before_action :authenticate_user!, only: %i(index)

    def index
      @channels = Channel.only_kept.select(:id)
      @received_channel_ids = current_user.receptions.pluck(:channel_id)
    end
  end
end
