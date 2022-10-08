# frozen_string_literal: true

class Deprecated::ChannelsQuery
  # @param collection [Channel::ActiveRecord_Relation]
  # @param is_vod [Boolean, nil]
  #
  # @return [Channel::ActiveRecord_Relation]
  def initialize(collection, is_vod: nil)
    @collection = collection
    @is_vod = is_vod
  end

  def call
    @collection = filter_by_is_vod unless is_vod.nil?
    @collection
  end

  private

  attr_reader :is_vod

  def filter_by_is_vod
    return @collection if is_vod.nil?

    @collection.where(vod: is_vod)
  end
end
