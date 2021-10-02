# frozen_string_literal: true

module Forms
  class CollectionItemForm < Forms::ApplicationForm
    attr_accessor :collection_id, :collection_item, :user, :work
    attr_reader :body

    validates :body, length: {maximum: 1_048_596}
    validates :collection, presence: true

    def body=(value)
      @body = value&.strip
    end

    def collection
      collection_item&.collection.presence || @user&.collections&.only_kept&.find_by(id: @collection_id)
    end

    # @overload
    def persisted?
      !collection_item.nil?
    end
  end
end
