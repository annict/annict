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

  # @param [Hash, nil] attributes
  def initialize(attributes = nil)
    @attributes = attributes

    if @attributes
      assign_attributes(@attributes)
    end
  end

  def valid?
    return true unless attributes

    @validation_result = contract.new.call(attributes)
    assign_attributes(@validation_result.to_h)

    !@validation_result.failure?
  end

  def error_messages
    return [] unless validation_result

    separator = I18n.locale == :ja ? "" : " "
    validation_result.errors.to_h.map do |attr_name, predicates|
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

  attr_reader :attributes, :validation_result

  def contract
    form_class_name = self.class.name
    contract_class_name = form_class_name.sub(%r{Form\z}, "Contract")

    ApplicationContract.const_get(contract_class_name)
  end
end
