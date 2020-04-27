# frozen_string_literal: true

module WorkDetail
  class TrailerEntity < ApplicationEntity
    attribute :title, Types::String
    attribute :url, Types::String
    attribute :image_url, Types::String
  end
end
