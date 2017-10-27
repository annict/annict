# frozen_string_literal: true

module Db
  class ChannelsController < Db::ApplicationController
    def index
      @channels = Channel.published.order(video_service: :desc, id: :desc)
    end
  end
end
