# frozen_string_literal: true

module ResourceImageDecoratorCommon
  extend ActiveSupport::Concern

  included do
    def asin_or_copyright_text
      if copyright.present?
        messages = []
        messages << icon("copyright", "far", class: "me-1")
        messages << Rack::Utils.escape_html(copyright)
        content_tag(:span, messages.join.html_safe, class: "text-muted")
      elsif asin.present?
        link_to amazon_url, target: "_blank", rel: "noopener" do
          messages = []
          messages << icon("amazon", "fab", class: "me-1")
          messages << I18n.t("messages._common.view_amazon_product")
          messages.join.html_safe
        end
      end
    end

    private

    def amazon_url
      amazon_url_key = I18n.locale == :ja ? "AMAZON_JA_URL" : "AMAZON_EN_URL"
      "#{ENV.fetch(amazon_url_key)}/dp/#{asin}"
    end
  end
end
