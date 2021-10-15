# frozen_string_literal: true

module Forms
  class NoteForm < Forms::ApplicationForm
    attr_accessor :library_entry
    attr_reader :body

    validates :body, length: {maximum: 1_048_596}
    validates :library_entry, presence: true

    def body=(value)
      @body = value&.strip
    end

    # @overload
    def persisted?
      true
    end
  end
end
