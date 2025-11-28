# typed: false
# frozen_string_literal: true

module Forms
  class ApplicationForm
    include ActiveModel::Model

    # @overload
    def self.i18n_scope
      :form
    end
  end
end
