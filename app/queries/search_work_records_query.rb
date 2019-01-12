# frozen_string_literal: true

class SearchWorkRecordsQuery
  def initialize(collection = WorkRecord.all, order_by: nil, has_body: nil)
    @collection = collection.published
    @args = {
      order_by: order_by,
      has_body: has_body
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
      when "LIKES_COUNT"
        @collection.order(likes_count: direction)
      end
    end

    @collection = case @args[:has_body]
    when true
      @collection.with_body
    when false
      @collection.with_no_body
    else
      @collection
    end

    @collection
  end
end
