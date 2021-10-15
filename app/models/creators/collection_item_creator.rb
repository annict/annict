# frozen_string_literal: true

module Creators
  class CollectionItemCreator
    def initialize(user:, form:)
      @user = user
      @form = form
    end

    def call
      @user.collection_items.create!(collection: @form.collection, work: @form.work)
    end
  end
end
