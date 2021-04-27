# frozen_string_literal: true

module Fragment
  class ReceiveChannelButtonsController < Fragment::ApplicationController
    before_action :authenticate_user!

    def index
      @channels = Channel.only_kept.select(:id)
      received_channel_ids = current_user.receptions.pluck(:channel_id)

      @channels.each do |channel|
        channel.is_received = received_channel_ids.include?(channel.id)
      end
    end
  end
end
