# typed: false
# frozen_string_literal: true

module Deprecated::Buttons
  class ShareToFacebookButtonComponent < Deprecated::ApplicationV6Component
    def initialize(view_context, url:, class_name: "")
      super view_context
      @url = url
      @class_name = class_name
    end

    def render
      build_html do |h|
        h.tag :span,
          class: "c-share-to-facebook-button #{@class_name}",
          data_controller: "share-to-facebook-button",
          data_share_to_facebook_button_url: @url,
          data_share_to_facebook_button_app_id: ENV.fetch("FACEBOOK_APP_ID") do
            h.tag :span, class: "btn btn-sm u-btn-facebook", data_action: "click->share-to-facebook-button#open" do
              h.tag :div, class: "small" do
                h.tag :i, class: "fab fa-facebook me-1"
                h.text t("noun.share")
              end
            end
          end
      end
    end
  end
end
