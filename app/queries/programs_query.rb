# frozen_string_literal: true

class ProgramsQuery
  # @param collection [Program::ActiveRecord_Relation]
  # @param order [GraphqlOrderStruct]
  #
  # @return [Program::ActiveRecord_Relation]
  def initialize(collection, order: GraphqlOrderStruct.new(:created_at, :asc))
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
