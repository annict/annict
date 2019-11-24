# frozen_string_literal: true

class ProgramsQuery
  # @param [Program::ActiveRecord_Relation] collection
  # @param [GraphqlOrderStruct] order
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
    return @collection unless @order

    case @order.field
    when "STARTED_AT"
      @collection.order(started_at: @order.direction)
    else
      @collection
    end
  end
end
