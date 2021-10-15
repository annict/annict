# frozen_string_literal: true

module Destroyers
  class CollectionDestroyer
    attr_accessor :user

    def initialize(collection:)
      @collection = collection
      @user = @collection.user
    end

    def call
      @collection.destroy!

      self
    end
  end
end
