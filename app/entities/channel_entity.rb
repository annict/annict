# frozen_string_literal: true

class ChannelEntity < ApplicationEntity
  attribute? :database_id, Types::Integer
  attribute? :name, Types::String
end
