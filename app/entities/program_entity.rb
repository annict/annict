# frozen_string_literal: true

class ProgramEntity < ApplicationEntity
  attribute? :vod_title_name, Types::String.optional
  attribute? :vod_title_url, Types::String.optional
  attribute? :channel, ChannelEntity
end
