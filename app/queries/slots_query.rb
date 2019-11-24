# frozen_string_literal: true

class SlotsQuery
  # @param collection [Slot::ActiveRecord_Relation]
  # @param order [OrderProperty]
  #
  # @return [Slot::ActiveRecord_Relation]
  def initialize(collection, order: OrderProperty.new)
    @collection = collection
    @order = order
  end

  def call
    order_collection
  end

  private

  attr_reader :order

  def order_collection
    return @collection.order(:created_at) unless order

    @collection.order(order.field => order.direction)
  end
end
