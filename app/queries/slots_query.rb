# frozen_string_literal: true

class SlotsQuery
  class OrderProperty
    def initialize(field_ = nil, direction_ = nil)
      @field_ = field_
      @direction_ = direction_
    end

    def field
      field_&.to_s&.downcase&.to_sym.presence || :created_at
    end

    def direction
      direction_&.to_s&.downcase&.to_sym.presence || :asc
    end

    private

    attr_reader :field_, :direction_
  end

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
