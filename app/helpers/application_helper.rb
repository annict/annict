module ApplicationHelper
  def annict_image_tag(source, options = {})
    options[:size] = if browser.mobile?
      options[:msize]
    else
      options[:psize]
    end

    image_tag(source, options)
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
end
