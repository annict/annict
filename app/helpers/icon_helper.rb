# typed: false
# frozen_string_literal: true

module IconHelper
  def icon(icon, style_prefix = "fas", html_options = {})
    if style_prefix.is_a?(Hash)
      html_options = style_prefix
      style_prefix = "fas"
    end

    content_class = "#{style_prefix} fa-#{icon}"
    content_class += " #{html_options[:class]}" if html_options.key?(:class)
    html_options[:class] = content_class

    content_tag(:i, nil, html_options)
  end
end
