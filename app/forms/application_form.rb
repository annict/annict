# frozen_string_literal: true

class ApplicationForm
  include ActiveModel::AttributeAssignment
  # Include to display form label name using `to_model`
  # https://github.com/rails/rails/blob/bdc581616b760d1e2be3795c6f0f3ab4b1e125a5/actionview/lib/action_view/helpers/tags/translator.rb#L9
  include ActiveModel::Conversion

  extend ActiveModel::Naming
  extend ActiveModel::Translation

  # @overload
  def self.i18n_scope
    :form
  end

  # @param [Dry::Validation::Result, nil] safe_params
  def initialize(safe_params = nil)
    @safe_params = safe_params

    if safe_params
      assign_attributes(safe_params.to_h)
    end
  end

  def valid?
    return true unless safe_params

    !safe_params.failure?
  end

  def error_messages
    return [] unless safe_params

    separator = I18n.locale == :ja ? "" : " "
    safe_params.errors.to_h.map do |attr_name, predicates|
      [
        self.class.human_attribute_name(attr_name),
        predicates.first
      ].join(separator)
    end
  end

  def persisted?
    false
  end

  private

  attr_reader :safe_params
end
