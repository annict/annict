# frozen_string_literal: true

class SlotsQuery
  # @param [Slot::ActiveRecord_Relation] collection
  # @param [GraphqlOrderStruct] order
  #
  # @return [Slot::ActiveRecord_Relation]
  def initialize(collection, order:)
    @collection = collection
    @order = order
  end

  def call
    @collection = order_collection if @order
    @collection
  end

  private

  def order_collection
    case @order.field
    when "STARTED_AT"
      @collection.order(started_at: @order.direction)
    else
      @collection
    end
  end
end
