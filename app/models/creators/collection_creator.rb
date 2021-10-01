# frozen_string_literal: true

module Creators
  class CollectionCreator
    attr_accessor :collection

    def initialize(user:, form:)
      @user = user
      @form = form
    end

    def call
      @collection = @user.collections.new(
        name: @form.name,
        description: @form.description
      )

      @collection.save!

      self
    end
  end
end
