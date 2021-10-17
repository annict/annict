# frozen_string_literal: true

class AdsenseComponent < ApplicationV6Component
  def initialize(view_context, slot:)
    super view_context
    @slot = slot
  end

  def render
    return "" if @slot.nil?
    return "" if current_user&.supporter?

    build_html do |h|
      h.tag :div, class: "container mt-3 text-center" do
        if display_ads?
          h.tag :ins, {
            class: "adsbygoogle d-block mt-1",
            data_ad_client: ENV.fetch("GOOGLE_AD_CLIENT"),
            data_ad_format: "horizontal",
            data_ad_slot: @slot,
            data_full_width_responsive: "false"
          }
        else
          h.tag :img, class: "img-fluid", src: "http://via.placeholder.com/300x100"
        end

        h.tag :div, class: "mt-1 small text-muted" do
          h.tag :i, class: "far fa-sparkles me-1 text-warning"
          h.html t("messages._components.adsense.hide_ads_html")
        end
      end
    end
  end
end
