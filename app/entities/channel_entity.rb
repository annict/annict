# frozen_string_literal: true

class ChannelEntity < ApplicationEntity
  attribute? :id, Types::Integer
  attribute? :name, Types::String
end
