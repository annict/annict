# frozen_string_literal: true

class TrailerObject < ApplicationObject
  attribute :title, ObjectTypes::Strict::String
  attribute :url, ObjectTypes::Strict::String
  attribute :internal_image_url, ObjectTypes::Strict::String
end
