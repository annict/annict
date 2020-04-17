# frozen_string_literal: true

class SearchSeriesWorksQuery
  def initialize(
    collection = SeriesWork.all,
    order_by: nil
  )
    @collection = collection.only_kept
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
      when "SEASON"
        @collection.sort_season(sort_type: direction)
      end
    end

    @collection
  end
end
