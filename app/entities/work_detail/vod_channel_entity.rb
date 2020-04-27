# frozen_string_literal: true

module WorkDetail
  class VodChannelEntity < ApplicationEntity
    attribute :id, Types::Integer
    attribute :name, Types::String
    attribute :programs, Types::Array.of(ProgramEntity)
  end
end
