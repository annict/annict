# frozen_string_literal: true

class SeriesWorkEntity < ApplicationEntity
  local_attributes :summary, :title

  attribute? :summary, Types::String.optional
  attribute? :summary_en, Types::String.optional
  attribute? :id, Types::Integer
  attribute? :title, Types::String
  attribute? :title_en, Types::String.optional
  attribute? :image_url_1x, Types::String.optional
  attribute? :image_url_2x, Types::String.optional
end
