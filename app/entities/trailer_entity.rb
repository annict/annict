# frozen_string_literal: true

class TrailerEntity < ApplicationEntity
  attribute? :title, Types::String
  attribute? :url, Types::String
  attribute? :image_url, Types::String
end
