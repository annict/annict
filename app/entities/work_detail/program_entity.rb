# frozen_string_literal: true

module WorkDetail
  class ProgramEntity < ApplicationEntity
    attribute :vod_title_name, Types::String.optional
    attribute :vod_title_url, Types::String.optional
    attribute :channel do
      attribute :id, Types::Integer
      attribute :name, Types::String
    end
  end
end
