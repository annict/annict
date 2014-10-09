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

  def meta_keywords(keywords = [])
    default_keywords = [
      'annict',
      'アニクト',
      'アニメ',
      '深夜アニメ',
      '視聴記録',
      '視聴管理',
      '見た',
      '観た',
      '記録',
      '管理',
      '感想',
      'sns',
      'ソーシャル'
    ]

    (keywords + default_keywords).join(',')
  end
end
