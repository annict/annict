# frozen_string_literal: true

class ApplicationForm
  include ActiveModel::Model

  # @overload
  def self.i18n_scope
    :form
  end
end
