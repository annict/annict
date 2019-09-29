# frozen_string_literal: true

class SearchChannelsRepository
  def initialize(
    collection = Channel.all,
    is_vod:
  )
    @collection = collection
    @args = {
      is_vod: is_vod
    }
  end

  def call
    from_arguments
  end

  private

  def from_arguments
    %i(
      is_vod
    ).each do |arg_name|
      next if @args[arg_name].nil?
      @collection = send(arg_name)
    end

    @collection
  end

  def is_vod
    @collection.where(vod: @args[:is_vod])
  end
end
