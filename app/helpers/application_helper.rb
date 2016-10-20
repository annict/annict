module ApplicationHelper
  def annict_image_url(record, field, options = {})
    path = record.try!(:send, field).try!(:path, :master).presence || "no-image.jpg"
    path = path.sub(%r{\A.*paperclip/}, "paperclip/") unless Rails.env.production?

    msize = options[:msize]
    size = (browser.mobile? && msize.present?) ? msize : options[:size]
    width, height = size.split("x").map { |s| s.to_i * 2 }

    blur = options[:blur].presence || 0

    ix_image_url(path, w: width, h: height, fit: "crop", auto: "format", blur: blur)
  end

  def annict_image_tag(record, field, options = {})
    url = annict_image_url(record, field, options)

    msize = options[:msize]
    options[:size] = msize if browser.mobile? && msize.present?
    options.delete(:msize) if options.key?(:msize)

    image_tag(url, options)
  end

  def custom_time_ago_in_words(datetime)
    days = (Time.zone.now.to_date - datetime.to_date).to_i

    if days > 3
      datetime.strftime("%Y/%m/%d")
    else
      "#{time_ago_in_words(datetime)}#{t('words.ago')}"
    end
  end

  def meta_description(text = "")
    text + t("meta.description")
  end

  def meta_keywords(*keywords)
    default_keywords = t("meta.keywords").split(",")
    (keywords + default_keywords).join(",")
  end

  # Google Analyticsのカスタムディメンション「ページカテゴリ」に送信する文字列
  def ga_page_category
    return "top" if top_page?
    return "programs" if programs_page?
    return "user_profile" if user_profile_page?
    return "user_works" if user_works_page?
    return "works_season" if works_season_page?
    return "works_popular" if works_popular_page?
    return "work_detail" if work_detail_page?
    return "search" if search_page?
    return "episode_detail" if episode_detail_page?
    return "person_detail" if person_detail_page?
    return "organization_detail" if organization_detail_page?
    "other"
  end

  def programs_page?
    params[:controller] == "programs" && params[:action] == "index"
  end

  def user_works_page?
    params[:controller] == "users" && params[:action] == "works"
  end

  def works_season_page?
    params[:controller] == "works" && params[:action] == "season"
  end

  def works_popular_page?
    params[:controller] == "works" && params[:action] == "popular"
  end

  def work_detail_page?
    params[:controller] == "works" && params[:action] == "show"
  end

  def search_page?
    params[:controller] == "searches" && params[:action] == "show"
  end

  def episode_detail_page?
    params[:controller] == "episodes" && params[:action] == "show"
  end

  def user_profile_page?
    params[:controller] == "users" && params[:action] == "show"
  end

  def top_page?
    params[:controller] == "home" && params[:action] == "index"
  end

  def person_detail_page?
    params[:controller] == "people" && params[:action] == "show"
  end

  def organization_detail_page?
    params[:controller] == "organizations" && params[:action] == "show"
  end

  def body_classes
    controller_name = controller.controller_path.tr("/", "-")
    basic_body_classes = [
      "p-#{controller_name}",
      "p-#{controller_name}-#{controller.action_name}"
    ].join(" ")

    if content_for?(:extra_body_classes)
      [basic_body_classes, content_for(:extra_body_classes)].join(" ")
    else
      basic_body_classes
    end
  end

  def locale_ja?
    locale == :ja || (user_signed_in? && current_user.role.admin?)
  end

  def v1_display_meta_tags
    display_meta_tags(
      site: "Annict",
      og: {
        title: page_title(page_title_symbol: :site_page_title, separator: " | "),
        type: "website",
        url: request.url,
        description: t("og.description"),
        site_name: t("words.site_name"),
        image: "#{ENV.fetch('ANNICT_URL')}/images/og_image.png",
        app_id: "602271853188285",
        locale: "ja_JP"
      }
    )
  end
end
