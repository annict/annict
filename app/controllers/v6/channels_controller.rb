# frozen_string_literal: true

module V6
  class ChannelsController < V6::ApplicationController
    def index
      @channel_groups = ChannelGroup.only_kept.order(:sort_number)
      @channels = Channel.only_kept.order(:channel_group_id, :sort_number, :id).select(:id, :channel_group_id, :name)
    end
  end
end
