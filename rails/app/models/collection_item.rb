# typed: false
# frozen_string_literal: true

class CollectionItem < ApplicationRecord
  include SoftDeletable

  acts_as_list scope: :collection_id

  counter_culture :collection, column_name: ->(collection_item) { collection_item.not_deleted? ? :collection_items_count : nil }

  belongs_to :user
  belongs_to :collection
  belongs_to :work
end
