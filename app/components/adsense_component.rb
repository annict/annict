# frozen_string_literal: true

class AdsenseComponent < ApplicationV6Component
  def initialize(view_context, slot:)
    super view_context
    @slot = slot
  end

  # standard:disable Lint/UnreachableCode
  def render
    return ""
    return "" unless @slot
    return "" if current_user&.supporter?

    build_html do |h|
      h.tag :div, class: "py-3 text-center" do
        if Rails.env.production?
          h.tag :ins, {
            class: "c-adsense adsbygoogle d-block",
            data_ad_client: ENV.fetch("GOOGLE_AD_CLIENT"),
            data_ad_format: "horizontal",
            data_ad_slot: @slot,
            data_full_width_responsive: "true"
          }
        else
          h.tag :img, class: "img-fluid", src: "http://via.placeholder.com/300x100"
        end

        h.tag :div, class: "mt-1 small text-muted" do
          h.tag :i, class: "far fa-info-circle me-1"
          h.html t("messages._components.adsense.hide_ads_html")
        end
      end
    end
  end
  # standard:enable Lint/UnreachableCode
end
