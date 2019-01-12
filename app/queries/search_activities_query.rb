# frozen_string_literal: true

class SearchActivitiesQuery
  def initialize(collection = Activity.all, order_by: nil)
    @collection = collection
    @args = {
      order_by: order_by
    }
  end

  def call
    from_arguments
  end

  private

  def from_arguments
    if @args[:order_by].present?
      direction = @args[:order_by][:direction]

      @collection = case @args[:order_by][:field]
      when "CREATED_AT"
        @collection.order(created_at: direction)
      end
    end

    @collection
  end
end
