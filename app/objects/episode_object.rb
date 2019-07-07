# frozen_string_literal: true

class EpisodeObject < ApplicationObject
  attribute :annict_id, ObjectTypes::Strict::Integer
  attribute :number_text, ObjectTypes::Strict::String
  attribute :title, ObjectTypes::Strict::String
end
