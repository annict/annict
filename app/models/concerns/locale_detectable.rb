# frozen_string_literal: true

module LocaleDetectable
  extend ActiveSupport::Concern

  included do
    enumerize :locale, in: ApplicationRecord::LOCALES, default: :other

    def detect_locale!(column)
      result = CLD.detect_language(send(column))
      self.locale = result[:code] if result[:code].in?(self.class.locale.values)
    end
  end
end
