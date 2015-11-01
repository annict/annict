module ApplicationHelper
  def annict_image_url(record, field, options = {})
    path = record.send(field).path(:master) rescue "no-image.jpg"

    msize = options[:msize]
    size = (browser.mobile? && msize.present?) ? msize : options[:size]
    width, height = size.split("x").map { |s| s.to_i * 2 }

    ix_image_url(path, w: width, h: height, fit: "crop")
  end

  def annict_image_tag(record, field, options = {})
    url = annict_image_url(record, field, options)

    msize = options[:msize]
    options[:size] = msize if browser.mobile? && msize.present?
    options.delete(:msize) if options.key?(:msize)

    image_tag(url, options)
  end

  def custom_time_ago_in_words(date)
    "#{time_ago_in_words(date)}#{t('words.ago')}"
  end

  def meta_description(text = '')
    text + t('meta.description')
  end

  def meta_keywords(*keywords)
    default_keywords = t('meta.keywords').split(',')
    (keywords + default_keywords).join(',')
  end

  def programs_page?
    params[:controller] == 'programs' && params[:action] == 'index'
  end

  def user_works_page?
    params[:controller] == 'users' && params[:action] == 'works'
  end

  def works_page?
    params[:controller] == 'works' &&
    (params[:action] == 'season' ||
     params[:action] == 'popular' ||
     params[:action] == 'recommend' ||
     params[:action] == 'search')
  end

  def user_profile_page?
    params[:controller] == 'users' && params[:action] == 'show'
  end
end
