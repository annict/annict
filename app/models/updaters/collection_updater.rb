# frozen_string_literal: true

module Updaters
  class CollectionUpdater
    attr_accessor :collection

    def initialize(user:, form:)
      @user = user
      @form = form
      @collection = @form.collection
    end

    def call
      @collection.name = @form.name
      @collection.description = @form.description

      @collection.save!

      self
    end
  end
end
