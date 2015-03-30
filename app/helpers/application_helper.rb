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

  def programs_page?
    params[:controller] == 'programs' && params[:action] == 'index'
  end

  def user_works_page?
    params[:controller] == 'users' && params[:action] == 'works'
  end

  def works_page?
    params[:controller] == 'works' &&
    (params[:action] == 'on_air' ||
     params[:action] == 'season' ||
     params[:action] == 'popular' ||
     params[:action] == 'recommend' ||
     params[:action] == 'search')
  end

  def user_profile_page?
    params[:controller] == 'users' && params[:action] == 'show'
  end

  def thumb_url(model, accessor_name, size)
    path = model.send("#{accessor_name}_uid")

    # プロフィール背景画像がGifアニメのときは、S3に保存された画像をそのまま返す
    if accessor_name == :background_image && model.background_image_animated?
      return "#{ENV['ANNICT_IMAGE2_URL']}/images/#{path}"
    end

    "#{ENV['ANNICT_IMAGE2_URL']}/#{size}/#{path}"
  end

  def tombo_thumb_url(model, accessor_name, size = "")
    image = model.send(accessor_name)

    # プロフィール背景画像がGifアニメのときは、S3に保存された画像をそのまま返す
    if accessor_name == :tombo_background_image && model.background_image_animated?
      return "#{ENV['ANNICT_FILE_STORAGE_URL']}/#{image.path(:original)}"
    end

    "#{ENV['ANNICT_TOMBO_URL']}/#{size}/#{image.path(:master)}"
  end
end
