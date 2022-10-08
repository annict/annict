# frozen_string_literal: true

class Deprecated::SearchStaffsQuery
  def initialize(collection = Staff.all, order_by: nil)
    @collection = collection.only_kept.preload(:resource)
    @args = {
      order_by: order_by
    }
  end

  def call
    from_arguments
  end

  private

  def from_arguments
    if @args[:order_by]
      direction = @args[:order_by][:direction]

      @collection = case @args[:order_by][:field]
      when "CREATED_AT"
        @collection.order(created_at: direction)
      when "SORT_NUMBER"
        @collection.order(sort_number: direction)
      end
    end

    @collection
  end
end
