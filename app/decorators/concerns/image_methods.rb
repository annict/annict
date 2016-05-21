# frozen_string_literal: true

module ImageMethods
  extend ActiveSupport::Concern

  included do
    def image_url(field, options = {})
      h.annict_image_url(self, field, options)
    end
  end
end
