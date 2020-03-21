# frozen_string_literal: true

class ChannelsController < ApplicationController
  before_action :authenticate_user!

  def index
    @channel_groups = ChannelGroup.only_kept.order(:sort_number)
  end
end
