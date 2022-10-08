# frozen_string_literal: true

module Deprecated::Buttons
  class ShareToTwitterButtonComponent < Deprecated::ApplicationV6Component
    def initialize(view_context, text:, url:, hashtags: "", class_name: "")
      super view_context
      @text = text
      @url = url
      @hashtags = hashtags
      @class_name = class_name
    end

    def render
      build_html do |h|
        h.tag :span,
          class: "c-share-to-twitter-button #{@class_name}",
          data_controller: "share-to-twitter-button",
          data_share_to_twitter_button_text: @text,
          data_share_to_twitter_button_url: @url,
          data_share_to_twitter_button_hashtags: @hashtags do
            h.tag :span, class: "btn btn-sm text-white u-btn-twitter", data_action: "click->share-to-twitter-button#open" do
              h.tag :div, class: "small" do
                h.tag :i, class: "fab fa-twitter me-1"
                h.text t("noun.tweet")
              end
            end
          end
      end
    end
  end
end
