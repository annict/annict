# frozen_string_literal: true

module Forms
  class CollectionForm < Forms::ApplicationForm
    attr_accessor :collection
    attr_reader :description, :name

    validates :name, presence: true, length: {maximum: 50}
    validates :description, length: {maximum: 1_048_596}

    def description=(value)
      @description = value&.strip
    end

    def name=(value)
      @name = value&.strip
    end

    # @overload
    def persisted?
      !collection.nil?
    end
  end
end
