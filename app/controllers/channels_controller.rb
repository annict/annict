class ChannelsController < ApplicationController
  before_filter :authenticate_user!

  def index
    @channel_groups = ChannelGroup.published.order(:sort_number)
  end
end