# frozen_string_literal: true

class SearchSlotsQuery
  def initialize(collection = Slot.all, order_by: nil, unwatched: nil)
    @collection = collection.published
    @args = {
      unwatched: unwatched,
      order_by: order_by
    }
  end

  def call
    from_arguments
  end

  private

  def from_arguments
    @collection = if @args[:unwatched].present?
      @collection.unwatched_all
    else
      @collection.all
    end

    @collection = @collection.work_published.episode_published

    if @args[:orderBy].present?
      direction = @args[:orderBy][:direction]

      @collection = case @args[:orderBy][:field]
      when "STARTED_AT"
        @collection.order(started_at: direction)
      end
    end

    @collection
  end
end
