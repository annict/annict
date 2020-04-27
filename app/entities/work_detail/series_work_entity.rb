# frozen_string_literal: true

module WorkDetail
  class SeriesWorkEntity < ApplicationEntity
    local_attributes :summary, :title

    attribute :summary, Types::String.optional
    attribute :summary_en, Types::String.optional
    attribute :id, Types::Integer
    attribute :title, Types::String
    attribute :title_en, Types::String.optional
    attribute :image_url, Types::String.optional
  end
end
