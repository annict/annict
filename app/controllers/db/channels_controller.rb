# frozen_string_literal: true

module Db
  class ChannelsController < Db::ApplicationController
    def index
      @channels = Channel.published.order(id: :desc)
    end
  end
end
