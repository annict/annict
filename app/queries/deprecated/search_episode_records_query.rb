# typed: false
# frozen_string_literal: true

class Deprecated::SearchEpisodeRecordsQuery
  def initialize(collection = EpisodeRecord.all, has_body: nil, order_by: nil)
    @collection = collection.only_kept
    @args = {
      has_comment: has_body,
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
      when "LIKES_COUNT"
        @collection.order(likes_count: direction)
      end
    end

    @collection = case @args[:has_comment]
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
