# frozen_string_literal: true

class SearchProgramsRepository
  def initialize(collection = Program.all, order_by:)
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
    if @args[:orderBy]
      direction = @args[:orderBy][:direction]

      @collection = case @args[:orderBy][:field]
      when "STARTED_AT"
        @collection.order(started_at: direction)
      end
    end

    @collection
  end
end
