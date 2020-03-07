# frozen_string_literal: true

module DB
  module WorkList
    class WorkConnectionEntity < ApplicationEntity
      attribute :works, Types::Array do
        attribute :id, Types::Integer
        attribute :title, Types::String
        attribute :title_kana, Types::String
        attribute :title_en, Types::String
        attribute :media_text, Types::String
        attribute :started_on, Types::String.optional
        attribute :syobocal_tid, Types::String.optional
        attribute :syobocal_url, Types::String.optional
        attribute :mal_anime_id, Types::String.optional
        attribute :mal_anime_url, Types::String.optional
        attribute :image_url, Types::String
        attribute :watchers_count, Types::Integer
        attribute :disappeared_at, Types::DateTime.optional
      end

      attribute :has_next_page, Types::Bool
      attribute :start_cursor, Types::String
    end
  end
end
