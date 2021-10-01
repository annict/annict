# frozen_string_literal: true

module Updaters
  class CollectionItemUpdater
    attr_accessor :collection_item

    def initialize(user:, form:)
      @user = user
      @form = form
      @collection_item = @form.collection_item
    end

    def call
      @collection_item.body = @form.body

      @collection_item.save!

      self
    end
  end
end
