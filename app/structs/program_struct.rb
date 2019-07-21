# frozen_string_literal: true

class ProgramStruct < ApplicationStruct
  attribute :annict_id, StructTypes::Strict::Integer
  attribute :vod_title_code, StructTypes::Strict::String
  attribute :vod_title_name, StructTypes::Strict::String

  attribute :channel, ChannelStruct
end
