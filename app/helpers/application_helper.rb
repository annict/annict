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

  def js_template(name)
    content_for :js_templates do
      render('js_template', name: name)
    end
  end
end
