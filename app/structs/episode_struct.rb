# frozen_string_literal: true

class EpisodeStruct < ApplicationStruct
  attribute :annict_id, StructTypes::Strict::Integer
  attribute :number_text, StructTypes::Strict::String
  attribute :title, StructTypes::Strict::String
end
