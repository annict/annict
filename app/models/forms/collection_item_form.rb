# frozen_string_literal: true

module Forms
  class CollectionItemForm < Forms::ApplicationForm
    attr_accessor :collection_item
    attr_reader :body

    validates :body, length: {maximum: 1_048_596}

    def body=(value)
      @body = value&.strip
    end

    # @overload
    def persisted?
      !collection_item.nil?
    end
  end
end
