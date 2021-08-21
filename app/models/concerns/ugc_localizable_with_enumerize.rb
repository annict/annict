# frozen_string_literal: true

module UgcLocalizableWithEnumerize
  extend ActiveSupport::Concern

  included do
    extend Enumerize

    enumerize :locale, in: ApplicationRecord::LOCALES, default: :other, scope: true

    scope :readable_by_user, ->(user) {
      with_locale(*user.allowed_locales).or(where(user: user))
    }

    def detect_locale!(column)
      result = CLD.detect_language(send(column))
      self.locale = result[:code] if result[:code].in?(self.class.locale.values)
    end
  end
end
