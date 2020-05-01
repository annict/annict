# frozen_string_literal: true

class EpisodeEntity < ApplicationEntity
  local_attributes :title

  attribute? :id, Types::Integer
  attribute? :number_text, Types::String.optional
  attribute? :title, Types::String.optional
  attribute? :title_en, Types::String.optional
end
