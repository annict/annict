# frozen_string_literal: true

module HeadHelper
  def ann_display_meta_tags
    display_meta_tags(
      reverse: true,
      site: "Annict",
      og: {
        title: page_title(page_title_symbol: :site_page_title, separator: " | "),
        type: "website",
        url: request.url,
        description: t("head.meta.description._common"),
        site_name: t("words.site_name"),
        image: "#{ENV.fetch('ANNICT_URL')}/images/og_image.png",
        locale: (I18n.locale == :ja ? "ja_JP" : "en_US")
      },
      fb: {
        app_id: ENV.fetch("FACEBOOK_APP_ID")
      },
      twitter: {
        card: "summary",
        site: "@anannict",
        title: "Annict",
        description: t("head.meta.description._common"),
        image: "#{ENV.fetch('ANNICT_URL')}/images/og_image.png"
      }
    )
  end

  def meta_description(text = "")
    "#{text} - #{t('head.meta.description._common')}"
  end

  def meta_keywords(*keywords)
    default_keywords = t("head.meta.keywords._common").split(",")
    (keywords + default_keywords).join(",")
  end
end
