# frozen_string_literal: true

module Destroyers
  class CollectionItemDestroyer
    attr_accessor :user

    def initialize(collection_item:)
      @collection_item = collection_item
      @user = @collection_item.user
    end

    def call
      @collection_item.destroy!

      self
    end
  end
end
