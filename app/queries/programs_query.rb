# frozen_string_literal: true

class ProgramsQuery
  # @param collection [Program::ActiveRecord_Relation]
  # @param order [GraphqlOrderStruct]
  #
  # @return [Program::ActiveRecord_Relation]
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
