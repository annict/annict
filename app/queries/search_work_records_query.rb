# frozen_string_literal: true

class SearchWorkRecordsQuery
  def initialize(collection = AnimeRecord.only_kept, context: {}, order_by: nil, has_body: nil, filter_by_locale: false)
    @collection = collection
    @context = context
    @args = {
      order_by: order_by,
      has_body: has_body,
      filter_by_locale: filter_by_locale
    }
  end

  def call
    from_arguments
  end

  private

  attr_reader :collection, :context, :args

  def from_arguments
    results = collection.joins(:user).merge(User.only_kept)
    viewer = context[:viewer]
    locale = context[:locale]

    if args[:order_by].present?
      direction = args[:order_by][:direction]

      results = case args[:order_by][:field]
      when "CREATED_AT"
        results.order(created_at: direction)
      when "LIKES_COUNT"
        results.order(likes_count: direction)
      end
    end

    results = case args[:has_body]
    when true
      results.with_body
    when false
      results.with_no_body
    else
      results
    end

    if locale && args[:filter_by_locale]
      results = if viewer
        results.readable_by_user(viewer)
      else
        results.with_locale(locale)
      end
    end

    results
  end
end
