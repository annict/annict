# frozen_string_literal: true

class ApplicationForm
  include ActiveModel::Model

  attr_writer :error_messages

  # @overload
  def self.i18n_scope
    :form
  end

  # @param [Hash, nil] attributes
  def initialize(attributes = nil)
    @attributes = attributes

    assign_attributes(@attributes) if @attributes
  end
end
