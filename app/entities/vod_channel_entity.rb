# frozen_string_literal: true

class VodChannelEntity < ApplicationEntity
  attribute? :id, Types::Integer
  attribute? :name, Types::String
  attribute? :programs, Types::Array.of(ProgramEntity)
end
