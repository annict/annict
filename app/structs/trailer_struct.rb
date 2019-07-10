# frozen_string_literal: true

class TrailerStruct < ApplicationStruct
  attribute :title, StructTypes::Strict::String
  attribute :url, StructTypes::Strict::String
  attribute :internal_image_url, StructTypes::Strict::String
end
