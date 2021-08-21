# frozen_string_literal: true

module UgcLocalizable
  extend ActiveSupport::Concern

  included do
    enum locale: ApplicationRecord::LOCALES, _prefix: true

    def detect_locale!(column)
      result = CLD.detect_language(send(column))
      self.locale = result[:code] if result[:code].in?(ApplicationRecord::LOCALES.map(&:to_s))
    end
  end
end
