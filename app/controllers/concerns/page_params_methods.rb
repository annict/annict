# frozen_string_literal: true

module PageParamsMethods
  extend ActiveSupport::Concern

  private

  def store_page_params(assigns)
    template_path = "application/page_params/#{page_category}"
    logger.info("store_page_params - template_path: #{template_path}") if Rails.env.development?
    return unless File.exist?(Rails.root.join("app/views/#{template_path}.jb"))
  end
end
