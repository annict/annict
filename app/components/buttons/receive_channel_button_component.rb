# frozen_string_literal: true

module Buttons
  class ReceiveChannelButtonComponent < ApplicationComponent
    def initialize(channel:, init_received: false, class_name: "")
      @channel = channel
      @init_received = init_received
      @class_name = class_name
    end

    private

    def receive_channel_button_class_name
      classes = %w(btn)
      classes += @class_name.split(" ")
      classes << (@init_received ? "btn-info" : "btn-outline-info")
      classes.uniq.join(" ")
    end

    def icon_name
      @init_received ? "fa-minus" : "fa-plus"
    end
  end
end
