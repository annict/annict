# frozen_string_literal: true

class LibraryEntriesQuery
  # @param library_entries [LibraryEntry::ActiveRecord_Relation]
  # @param filter_by [Hash]
  # @param order [OrderProperty]
  #
  # @return [LibraryEntry::ActiveRecord_Relation]
  def initialize(
    library_entries,
    filter_by:,
    order: OrderProperty.new
  )
    @library_entries = library_entries
    @filter_by = filter_by
    @order = order
  end

  def call
    collection = library_entries
    collection = filter_collection(collection)
    order_collection(collection)
  end

  private

  attr_reader :library_entries, :filter_by, :order

  def filter_collection(collection)
    return collection unless filter_by
    return collection if filter_by[:status_kinds].blank?

    status_kinds = filter_by[:status_kinds].map { |kind| Status.kind_v3_to_v2(kind.downcase) }

    collection.with_status(status_kinds)
  end

  def order_collection(collection)
    return collection.order(:created_at) unless order

    collection.order(order.field => order.direction)
  end
end
