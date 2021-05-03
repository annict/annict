# frozen_string_literal: true

module HeadHelper
  def ann_display_meta_tags(site = "Annict")
    display_meta_tags(
      reverse: true,
      site: site,
      separator: " |",
      description: meta_description,
      keywords: meta_keywords,
      og: {
        title: meta_tags.full_title(site: site, separator: " |"),
        type: "website",
        url: request.url,
        description: t("head.meta.description._common"),
        site_name: t("noun.annict"),
        image: "#{ENV.fetch("ANNICT_URL")}/images/og_image.png",
        locale: (I18n.locale == :ja ? "ja_JP" : "en_US")
      },
      fb: {
        app_id: ENV.fetch("FACEBOOK_APP_ID")
      },
      twitter: {
        card: "summary",
        site: twitter_username,
        title: meta_tags.full_title(site: site, separator: " |"),
        description: t("head.meta.description._common"),
        image: "#{ENV.fetch("ANNICT_URL")}/images/og_image.png"
      },
      "turbo-cache-control": "no-cache"
    )
  end

  def meta_description(text = "")
    ary = []
    ary << "#{text} -" if text.present?
    ary << t("head.meta.description._common")
    ary.join(" ")
  end

  def meta_keywords(*keywords)
    default_keywords = t("head.meta.keywords._common").split(",")
    (keywords + default_keywords).join(",")
  end
end
