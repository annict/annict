# frozen_string_literal: true

class ChannelsController < ApplicationController
  before_action :authenticate_user!

  def index
    @channel_groups = ChannelGroup.without_deleted.order(:sort_number)
  end
end
