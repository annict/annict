# frozen_string_literal: true

module PageParamsMethods
  extend ActiveSupport::Concern

  private

  def store_page_params(assigns)
    template_path = "application/page_params/#{page_category}"

    if Rails.env.development?
      logger.info("store_page_params - template_path: #{template_path}")
    end

    return unless File.exist?(Rails.root.join("app/views/#{template_path}.jb"))

    page = gon.page.presence || {}
    page[:params] = render_jb(template_path, assigns)
    gon.push(page: page)
  end
end
