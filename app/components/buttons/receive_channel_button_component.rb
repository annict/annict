# frozen_string_literal: true

module Buttons
  class ReceiveChannelButtonComponent < ApplicationComponent2
    def initialize(view_context, channel:, class_name: "")
      super view_context
      @channel = channel
      @class_name = class_name
      @init_received = @channel.is_received == true
    end

    def render
      build_html do |h|
        h.tag :button,
          class: "c-receive-channel-button #{receive_channel_button_class_name}",
          data_action: "click->receive-channel-button#toggle",
          data_controller: "receive-channel-button",
          data_receive_channel_button_channel_id_value: @channel.id,
          data_receive_channel_button_init_received_value: @init_received,
          data_receive_channel_button_not_received_button_class: "btn-outline-info",
          data_receive_channel_button_received_button_class: "btn-info" do
            h.tag :span, data_receive_channel_button_target: "iconWrapper" do
              h.tag :i, class: "fal #{icon_name}"
            end
          end
      end
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
