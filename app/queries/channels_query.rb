# frozen_string_literal: true

class ChannelsQuery
  # @param [Channel::ActiveRecord_Relation] collection
  # @param [Boolean] is_vod
  #
  # @return [Channel::ActiveRecord_Relation]
  def initialize(collection, is_vod:)
    @collection = collection
    @is_vod = is_vod
  end

  def call
    @collection = filter_by_is_vod if @is_vod
    @collection
  end

  private

  def filter_by_is_vod
    @collection.where(vod: @is_vod)
  end
end
